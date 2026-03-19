package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parseMemoryValue - парсинг значения памяти в MiB
func parseMemoryValue(val string) float64 {
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "n/a" || val == "" || val == "<none>" || val == "н/д" {
		return 0
	}

	// Парсим с разными суффиксами
	if strings.Contains(val, "gi") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "gi", "", -1), 64)
		return num * MiBInGiB
	}
	if strings.Contains(val, "mi") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "mi", "", -1), 64)
		return num
	}
	if strings.Contains(val, "ki") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "ki", "", -1), 64)
		return num / MiBInGiB
	}
	if strings.Contains(val, "g") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "g", "", -1), 64)
		return num * MiBInGiB
	}
	if strings.Contains(val, "m") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "m", "", -1), 64)
		return num / MiBInGiB / MiBInGiB
	}

	num, _ := strconv.ParseFloat(val, 64)
	return num
}

// parseCPUValue - парсинг значения CPU в millicores
func parseCPUValue(val string) float64 {
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "n/a" || val == "" || val == "<none>" || val == "н/д" {
		return 0
	}

	if strings.Contains(val, "m") {
		num, _ := strconv.ParseFloat(strings.Replace(val, "m", "", -1), 64)
		return num
	}

	num, _ := strconv.ParseFloat(val, 64)
	return num * MillicoresInCore
}

// formatCPUValue - форматирование значения CPU для вывода
func formatCPUValue(millicores float64) string {
	if millicores <= 0 {
		return "Н/Д"
	}
	return fmt.Sprintf("%.0fm", millicores)
}

// formatMemoryValue - форматирование значения памяти для вывода
func formatMemoryValue(mib float64) string {
	if mib <= 0 {
		return "Н/Д"
	}
	if mib >= MiBInGiB {
		return fmt.Sprintf("%.1fGi", mib/MiBInGiB)
	}
	return fmt.Sprintf("%.0fMi", mib)
}

// sanitizeSheetName - очистка имени листа Excel от недопустимых символов
func sanitizeSheetName(name string) string {
	if len(name) > 31 {
		name = name[:31]
	}
	re := regexp.MustCompile(`[\[\]:\*?/\\]`)
	return re.ReplaceAllString(name, "-")
}

// centerText - центрирование текста для консольного вывода
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

// parseSelection - парсинг выбора неймспейсов пользователем
func parseSelection(input string, maxIndex int) []int {
	var indices []int
	seen := make(map[int]bool)

	// Проверяем на диапазон (например: 1-5)
	if strings.Contains(input, "-") {
		parts := strings.Split(input, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil && start > 0 && end <= maxIndex && start <= end {
				for i := start; i <= end; i++ {
					if !seen[i] {
						seen[i] = true
						indices = append(indices, i)
					}
				}
			}
		}
	} else {
		// Разбираем числа через запятую
		parts := strings.FieldsFunc(input, func(r rune) bool {
			return r == ',' || r == ' '
		})
		for _, part := range parts {
			num, err := strconv.Atoi(strings.TrimSpace(part))
			if err == nil && num > 0 && num <= maxIndex {
				if !seen[num] {
					seen[num] = true
					indices = append(indices, num)
				}
			}
		}
	}

	return indices
}

// parsePodResourceFromJSON - парсинг ресурсов пода из JSON
func parsePodResourceFromJSON(item interface{}, namespace string) *PodResource {
	pod, ok := item.(map[string]interface{})
	if !ok {
		return nil
	}

	metadata, ok := pod["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}
	podName, _ := metadata["name"].(string)

	spec, ok := pod["spec"].(map[string]interface{})
	if !ok {
		return nil
	}

	nodeName := "неизвестно"
	if specNode, ok := spec["nodeName"]; ok {
		nodeName, _ = specNode.(string)
	}

	// Получаем привязанные PVC
	var pvcs []string
	if volumes, ok := spec["volumes"].([]interface{}); ok {
		for _, v := range volumes {
			vol, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			if pvc, ok := vol["persistentVolumeClaim"]; ok {
				pvcMap, ok := pvc.(map[string]interface{})
				if !ok {
					continue
				}
				if claimName, ok := pvcMap["claimName"]; ok {
					pvcs = append(pvcs, claimName.(string))
				}
			}
		}
	}

	// Парсим ресурсы контейнеров
	containers, ok := spec["containers"].([]interface{})
	if !ok || len(containers) == 0 {
		return nil
	}

	var totalCPUReq, totalCPULim, totalMemReq, totalMemLim float64
	for _, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		if resources, ok := container["resources"].(map[string]interface{}); ok {
			if requests, ok := resources["requests"].(map[string]interface{}); ok {
				if val, ok := requests["cpu"]; ok {
					totalCPUReq += parseCPUValue(val.(string))
				}
				if val, ok := requests["memory"]; ok {
					totalMemReq += parseMemoryValue(val.(string))
				}
			}
			if limits, ok := resources["limits"].(map[string]interface{}); ok {
				if val, ok := limits["cpu"]; ok {
					totalCPULim += parseCPUValue(val.(string))
				}
				if val, ok := limits["memory"]; ok {
					totalMemLim += parseMemoryValue(val.(string))
				}
			}
		}
	}

	return &PodResource{
		Namespace:     namespace,
		Name:          podName,
		NodeName:      nodeName,
		CPURequest:    formatCPUValue(totalCPUReq),
		CPULimit:      formatCPUValue(totalCPULim),
		MemoryRequest: formatMemoryValue(totalMemReq),
		MemoryLimit:   formatMemoryValue(totalMemLim),
		CPUActual:     "Н/Д",
		MemoryActual:  "Н/Д",
		PVCs:          pvcs,
	}
}

// parsePVCFromJSON - парсинг PVC из JSON
func parsePVCFromJSON(item interface{}) *PVCInfo {
	pvcMap, ok := item.(map[string]interface{})
	if !ok {
		return nil
	}

	metadata, ok := pvcMap["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	spec, ok := pvcMap["spec"].(map[string]interface{})
	if !ok {
		return nil
	}

	status, ok := pvcMap["status"].(map[string]interface{})
	if !ok {
		return nil
	}

	namespace, _ := metadata["namespace"].(string)
	name, _ := metadata["name"].(string)

	volumeName := ""
	if v, ok := spec["volumeName"]; ok {
		volumeName, _ = v.(string)
	}

	capacity := "0"
	if capMap, ok := status["capacity"].(map[string]interface{}); ok {
		if storage, ok := capMap["storage"]; ok {
			capacity, _ = storage.(string)
		}
	}

	storageClass := ""
	if sc, ok := spec["storageClassName"]; ok {
		storageClass, _ = sc.(string)
	}

	var accessModes []string
	if modes, ok := spec["accessModes"].([]interface{}); ok {
		for _, m := range modes {
			if mode, ok := m.(string); ok {
				accessModes = append(accessModes, mode)
			}
		}
	}

	phase, _ := status["phase"].(string)

	return &PVCInfo{
		Namespace:    namespace,
		Name:         name,
		Status:       phase,
		Volume:       volumeName,
		Capacity:     capacity,
		Requested:    capacity,
		Used:         "0",
		UsedPercent:  0.0,
		StorageClass: storageClass,
		AccessModes:  accessModes,
	}
}

// parsePVFromJSON - парсинг PV из JSON
func parsePVFromJSON(item interface{}) *PVInfo {
	pvMap, ok := item.(map[string]interface{})
	if !ok {
		return nil
	}

	metadata, ok := pvMap["metadata"].(map[string]interface{})
	if !ok {
		return nil
	}

	spec, ok := pvMap["spec"].(map[string]interface{})
	if !ok {
		return nil
	}

	status, ok := pvMap["status"].(map[string]interface{})
	if !ok {
		return nil
	}

	name, _ := metadata["name"].(string)

	capacity := "0"
	if capMap, ok := spec["capacity"].(map[string]interface{}); ok {
		if storage, ok := capMap["storage"]; ok {
			capacity, _ = storage.(string)
		}
	}

	storageClass := ""
	if sc, ok := spec["storageClassName"]; ok {
		storageClass, _ = sc.(string)
	}

	claimRef := ""
	if claim, ok := spec["claimRef"].(map[string]interface{}); ok {
		claimNamespace, _ := claim["namespace"].(string)
		claimName, _ := claim["name"].(string)
		claimRef = fmt.Sprintf("%s/%s", claimNamespace, claimName)
	}

	phase, _ := status["phase"].(string)

	return &PVInfo{
		Name:         name,
		Capacity:     capacity,
		Used:         "0",
		UsedPercent:  0.0,
		Status:       phase,
		Claim:        claimRef,
		StorageClass: storageClass,
	}
}
