package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func getStyleByEfficiency(styles map[string]int, eff float64) int {
	switch {
	case eff > EfficiencyCritical:
		return styles["critical"]
	case eff > EfficiencyHigh:
		return styles["high"]
	case eff > EfficiencyNormal:
		return styles["normal"]
	case eff > 0:
		return styles["low"]
	default:
		return styles["normal"]
	}
}

// generateExcelReport - генерация Excel отчёта кластера.
func generateExcelReport(cluster *ClusterSummary) string {
	printStep("📁 Создаем Excel отчет...")

	f := excelize.NewFile()
	defer f.Close()

	styles := createEnhancedStyles(f)

	createEnhancedClusterSummarySheetWithCharts(f, cluster, styles)
	createEnhancedNodesSheet(f, cluster, styles)
	createEnhancedPVCSheet(f, cluster, styles)
	createEnhancedPVSheet(f, cluster, styles)

	allPods := collectAllPodsForReport(cluster)
	createEnhancedNamespaceSheetsWithPods(f, cluster, allPods, styles)

	createGatekeeperSheet(f, cluster, styles)
	createRBACSheet(f, cluster, styles)

	if len(podHistories) > 0 {
		createHistorySheet(f, podHistories, styles)
	}

	// Sheet1 создаётся автоматически excelize — удаляем его
	if index, err := f.GetSheetIndex("Sheet1"); err == nil && index >= 0 {
		f.DeleteSheet("Sheet1")
	}

	f.SetActiveSheet(0)

	filename := fmt.Sprintf("k8s-анализ-%s.xlsx",
		time.Now().Format("2006-01-02-150405"))
	if err := f.SaveAs(filename); err != nil {
		printError(fmt.Sprintf("❌ Ошибка сохранения: %v", err))
		os.Exit(1)
	}

	printSuccess("   ✅ Excel отчет создан")
	return filename
}

func createEnhancedStyles(f *excelize.File) map[string]int {
	styles := make(map[string]int)

	mainHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#5B4A8F"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#FFFFFF", Style: 2},
			{Type: "right", Color: "#FFFFFF", Style: 2},
			{Type: "top", Color: "#FFFFFF", Style: 2},
			{Type: "bottom", Color: "#FFFFFF", Style: 2},
		},
	})
	styles["mainHeader"] = mainHeader

	tableHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#6B5B95"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#FFFFFF", Style: 1},
			{Type: "right", Color: "#FFFFFF", Style: 1},
			{Type: "top", Color: "#FFFFFF", Style: 1},
			{Type: "bottom", Color: "#FFFFFF", Style: 1},
		},
	})
	styles["tableHeader"] = tableHeader

	critical, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E74C3C"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["critical"] = critical

	high, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#000000",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F39C12"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["high"] = high

	normal, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#27AE60"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["normal"] = normal

	low, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#000000",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F4D03F"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["low"] = low

	data, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#D3D3D3", Style: 1},
			{Type: "right", Color: "#D3D3D3", Style: 1},
			{Type: "top", Color: "#D3D3D3", Style: 1},
			{Type: "bottom", Color: "#D3D3D3", Style: 1},
		},
	})
	styles["data"] = data

	number, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   12,
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFFFF"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#D3D3D3", Style: 1},
			{Type: "right", Color: "#D3D3D3", Style: 1},
			{Type: "top", Color: "#D3D3D3", Style: 1},
			{Type: "bottom", Color: "#D3D3D3", Style: 1},
		},
	})
	styles["number"] = number

	good, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#1E8449",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D5F4E6"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["good"] = good

	warning, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#9C5700",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FCE4C8"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})
	styles["warning"] = warning

	return styles
}


func createEnhancedNodesSheet(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "🖥️ Ноды"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "🖥️ АНАЛИЗ ЗАГРУЗКИ НОД КЛАСТЕРА")
	f.SetCellStyle(sheetName, "A1", "M1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "M1")
	f.SetRowHeight(sheetName, 1, 30)

	headers := []string{
		"Нода", "Поды",
		"CPU\nЕмкость", "CPU\nЗапросы", "CPU\nФакт", "CPU Зап\n%", "CPU Факт\n%",
		"Память\nЕмкость", "Память\nЗапросы", "Память\nФакт", "Пам Зап\n%", "Пам Факт\n%",
		"Рекомендации",
	}

	row := 3
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 40)
	row++

	var nodes []*NodeInfo
	for _, node := range cluster.ByNode {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	for _, node := range nodes {
		col := 'A'

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col, row), node.Name)
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col, row),
			fmt.Sprintf("%c%d", col, row), styles["data"])

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+1, row), node.PodsCount)
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+1, row),
			fmt.Sprintf("%c%d", col+1, row), styles["number"])

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+2, row),
			fmt.Sprintf("%.0fm", node.CPUCapacity))
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+3, row),
			fmt.Sprintf("%.0fm", node.CPURequests))
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+4, row),
			fmt.Sprintf("%.0fm", node.CPUActual))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+5, row),
			fmt.Sprintf("%.1f%%", node.CPURequestUtil))
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+5, row),
			fmt.Sprintf("%c%d", col+5, row),
			getStyleByEfficiency(styles, node.CPURequestUtil))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+6, row),
			fmt.Sprintf("%.1f%%", node.CPUUtilization))
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+6, row),
			fmt.Sprintf("%c%d", col+6, row),
			getStyleByEfficiency(styles, node.CPUUtilization))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+7, row),
			formatMemoryValue(node.MemoryCapacity))
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+8, row),
			formatMemoryValue(node.MemoryRequests))
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+9, row),
			formatMemoryValue(node.MemoryActual))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+10, row),
			fmt.Sprintf("%.1f%%", node.MemoryRequestUtil))
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+10, row),
			fmt.Sprintf("%c%d", col+10, row),
			getStyleByEfficiency(styles, node.MemoryRequestUtil))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+11, row),
			fmt.Sprintf("%.1f%%", node.MemoryUtilization))
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+11, row),
			fmt.Sprintf("%c%d", col+11, row),
			getStyleByEfficiency(styles, node.MemoryUtilization))

		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+12, row), node.Recommendation)
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+12, row),
			fmt.Sprintf("%c%d", col+12, row), styles["data"])

		row++
	}

	colWidths := []float64{30, 10, 15, 15, 15, 12, 12, 18, 18, 18, 12, 12, 50}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	lastCol := string(rune('A' + len(headers) - 1))
	f.AutoFilter(sheetName, fmt.Sprintf("A3:%s3", lastCol), nil)

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      1,
		YSplit:      3,
		TopLeftCell: "B4",
		ActivePane:  "bottomRight",
	})
}

// createEnhancedPVCSheet - создание улучшенного листа PVC
func createEnhancedPVCSheet(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "💽 PVC"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "💽 АНАЛИЗ PERSISTENT VOLUME CLAIMS")
	f.SetCellStyle(sheetName, "A1", "H1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "H1")
	f.SetRowHeight(sheetName, 1, 30)

	headers := []string{
		"Неймспейс", "Имя PVC", "Статус", "Том",
		"Емкость", "Запрошено", "Класс хранилища", "Режимы доступа",
	}

	row := 3
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 25)
	row++

	var pvcList []*PVCInfo
	for _, pvc := range cluster.ByPVC {
		pvcList = append(pvcList, pvc)
	}
	sort.Slice(pvcList, func(i, j int) bool {
		if pvcList[i].Namespace == pvcList[j].Namespace {
			return pvcList[i].Name < pvcList[j].Name
		}
		return pvcList[i].Namespace < pvcList[j].Namespace
	})

	for _, pvc := range pvcList {
		col := 'A'
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col, row), pvc.Namespace)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+1, row), pvc.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+2, row), pvc.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+3, row), pvc.Volume)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+4, row), pvc.Capacity)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+5, row), pvc.Requested)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+6, row), pvc.StorageClass)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+7, row),
			strings.Join(pvc.AccessModes, ", "))

		statusStyle := styles["good"]
		if pvc.Status != "Bound" {
			statusStyle = styles["warning"]
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
			fmt.Sprintf("%c%d", col+2, row), statusStyle)

		row++
	}

	colWidths := []float64{25, 35, 15, 35, 15, 15, 25, 30}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	f.AutoFilter(sheetName, "A3:H3", nil)

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})
}

// createEnhancedPVSheet - создание улучшенного листа PV
func createEnhancedPVSheet(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "💿 PV"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "💿 АНАЛИЗ PERSISTENT VOLUMES")
	f.SetCellStyle(sheetName, "A1", "E1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "E1")
	f.SetRowHeight(sheetName, 1, 30)

	headers := []string{
		"Имя PV", "Емкость", "Статус", "Привязка", "Класс хранилища",
	}

	row := 3
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 25)
	row++

	var pvList []*PVInfo
	for _, pv := range cluster.ByPV {
		pvList = append(pvList, pv)
	}
	sort.Slice(pvList, func(i, j int) bool {
		return pvList[i].Name < pvList[j].Name
	})

	for _, pv := range pvList {
		col := 'A'
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col, row), pv.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+1, row), pv.Capacity)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+2, row), pv.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+3, row), pv.Claim)
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+4, row), pv.StorageClass)

		statusStyle := styles["good"]
		if pv.Status != "Bound" {
			statusStyle = styles["warning"]
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
			fmt.Sprintf("%c%d", col+2, row), statusStyle)

		row++
	}

	colWidths := []float64{40, 15, 15, 45, 25}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	f.AutoFilter(sheetName, "A3:E3", nil)

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})
}


func collectAllPodsForReport(cluster *ClusterSummary) map[string]map[string]*PodResource {
	allPods := make(map[string]map[string]*PodResource)

	for ns := range cluster.ByNamespace {
		pods := getPodResources(ns)
		if len(pods) > 0 {
			if allPods[ns] == nil {
				allPods[ns] = make(map[string]*PodResource)
			}
			for _, pod := range pods {
				allPods[ns][pod.Name] = pod
			}
		}
		actualData := getPodActualUsage(ns)
		for podName, usage := range actualData {
			if pod, exists := allPods[ns][podName]; exists {
				pod.CPUActual = usage["cpu"]
				pod.MemoryActual = usage["memory"]
			}
		}
	}
	
	return allPods
}

// createEnhancedNamespaceSheetsWithPods - создание детальных листов неймспейсов с подами.
func createEnhancedNamespaceSheetsWithPods(f *excelize.File, cluster *ClusterSummary,
	allPods map[string]map[string]*PodResource, styles map[string]int) {

	printStep("   Создаем детальные листы неймспейсов...")

	var namespaces []string
	for ns := range cluster.ByNamespace {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	
	for _, ns := range namespaces {
		summary := cluster.ByNamespace[ns]
		if summary.PodCount == 0 {
			continue
		}
		
		sheetName := sanitizeSheetName(ns)
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		
		f.NewSheet(sheetName)

		title := fmt.Sprintf("📊 НЕЙМСПЕЙС: %s", ns)
		f.SetCellValue(sheetName, "A1", title)
		f.SetCellStyle(sheetName, "A1", "N1", styles["mainHeader"])
		f.MergeCell(sheetName, "A1", "N1")
		f.SetRowHeight(sheetName, 1, 35)
		
		row := 3

		row = addSectionHeader(f, sheetName, row, "📊 СТАТИСТИКА НЕЙМСПЕЙСА", styles)

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего подов:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), summary.PodCount)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
		
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "CPU Запрошено:")
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), 
			fmt.Sprintf("%.0fm (%.2f ядер)", summary.CPURequestTotal, 
			summary.CPURequestTotal/MillicoresInCore))
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles["good"])
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles["data"])
		
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Память Запрошено:")
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), formatMemoryValue(summary.MemRequestTotal))
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styles["good"])
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styles["data"])
		row++

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "CPU Фактически:")
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), 
			fmt.Sprintf("%.0fm (%.2f ядер)", summary.CPUActualTotal,
			summary.CPUActualTotal/MillicoresInCore))
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles["warning"])
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles["data"])
		
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), "Память Фактически:")
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), formatMemoryValue(summary.MemActualTotal))
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styles["warning"])
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styles["data"])
		row++

		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), 
			fmt.Sprintf("CPU Рекомендуется (+%d%%):", bufferPercent))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), 
			fmt.Sprintf("%.0fm (%.2f ядер)", summary.CPURecommendedTotal,
			summary.CPURecommendedTotal/MillicoresInCore))
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles["good"])
		
		cpuEffStyle := getStyleByEfficiency(styles, 
			(summary.CPUActualTotal/summary.CPURequestTotal)*100)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), cpuEffStyle)
		
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), 
			fmt.Sprintf("Память Рекомендуется (+%d%%):", bufferPercent))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), 
			formatMemoryValue(summary.MemRecommendedTotal))
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styles["good"])
		
		memEffStyle := getStyleByEfficiency(styles,
			(summary.MemActualTotal/summary.MemRequestTotal)*100)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), memEffStyle)
		row += 2

		row = addSectionHeader(f, sheetName, row, "📦 ДЕТАЛЬНАЯ ИНФОРМАЦИЯ ПО ПОДАМ", styles)

		headers := []string{
			"Имя пода", "Нода", "PVC",
			"CPU\nЗапрос", "CPU\nФакт", "CPU\nЛимит", "CPU\nЭфф %",
			"Память\nЗапрос", "Память\nФакт", "Память\nЛимит", "Память\nЭфф %",
			"Рекомендации", "Статус",
		}
		
		for i, h := range headers {
			cell := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue(sheetName, cell, h)
			f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
		}
		f.SetRowHeight(sheetName, row, 35)
		row++

		pods := allPods[ns]
		var podList []*PodResource
		for _, pod := range pods {
			podList = append(podList, pod)
		}
		sort.Slice(podList, func(i, j int) bool {
			return podList[i].Name < podList[j].Name
		})
		
		for _, pod := range podList {
			memEff := calculateMemoryEfficiency(pod)
			cpuEff := calculateCPUEfficiency(pod)
			recommendation := generatePodRecommendation(pod, memEff, cpuEff)
			status := getStatus(memEff)

			col := 'A'

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col, row), pod.Name)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col, row),
				fmt.Sprintf("%c%d", col, row), styles["data"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+1, row), pod.NodeName)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+1, row),
				fmt.Sprintf("%c%d", col+1, row), styles["data"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+2, row),
				strings.Join(pod.PVCs, ", "))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
				fmt.Sprintf("%c%d", col+2, row), styles["data"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+3, row), pod.CPURequest)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+3, row),
				fmt.Sprintf("%c%d", col+3, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+4, row), pod.CPUActual)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+4, row),
				fmt.Sprintf("%c%d", col+4, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+5, row), pod.CPULimit)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+5, row),
				fmt.Sprintf("%c%d", col+5, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+6, row),
				fmt.Sprintf("%.1f%%", cpuEff))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+6, row),
				fmt.Sprintf("%c%d", col+6, row), getStyleByEfficiency(styles, cpuEff))

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+7, row), pod.MemoryRequest)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+7, row),
				fmt.Sprintf("%c%d", col+7, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+8, row), pod.MemoryActual)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+8, row),
				fmt.Sprintf("%c%d", col+8, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+9, row), pod.MemoryLimit)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+9, row),
				fmt.Sprintf("%c%d", col+9, row), styles["number"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+10, row),
				fmt.Sprintf("%.1f%%", memEff))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+10, row),
				fmt.Sprintf("%c%d", col+10, row), getStyleByEfficiency(styles, memEff))

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+11, row), recommendation)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+11, row),
				fmt.Sprintf("%c%d", col+11, row), styles["data"])

			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+12, row), status)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+12, row),
				fmt.Sprintf("%c%d", col+12, row), getStyleByEfficiency(styles, memEff))

			row++
		}

		colWidths := []float64{40, 25, 30, 15, 15, 15, 12, 18, 18, 18, 12, 80, 18}
		for i, width := range colWidths {
			colName := string(rune('A' + i))
			f.SetColWidth(sheetName, colName, colName, width)
		}

		lastCol := string(rune('A' + len(headers) - 1))
		headerRow := row - len(podList) - 1
		f.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastCol, headerRow), nil)

		f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			XSplit:      1,
			YSplit:      headerRow,
			TopLeftCell: fmt.Sprintf("B%d", headerRow+1),
			ActivePane:  "bottomRight",
		})
	}
}

func addChartToSheet(f *excelize.File, sheetName string,
	chartType excelize.ChartType, title string,
	categories string, values string, 
	position string) error {
	
	if err := f.AddChart(sheetName, position, &excelize.Chart{
		Type: chartType,
		Series: []excelize.ChartSeries{
			{
				Name:       title,
				Categories: categories,
				Values:     values,
			},
		},
		Format: excelize.GraphicOptions{
			ScaleX:          1.0,
			ScaleY:          1.0,
			OffsetX:         15,
			OffsetY:         10,
			PrintObject:     boolPtr(true),
			LockAspectRatio: false,
			Locked:          boolPtr(false),
		},
		Title: []excelize.RichTextRun{
			{
				Text: title,
			},
		},
		PlotArea: excelize.ChartPlotArea{
			ShowCatName:     false,
			ShowLeaderLines: false,
			ShowPercent:     true,
			ShowSerName:     true,
			ShowVal:         true,
		},
		ShowBlanksAs: "zero",
		XAxis: excelize.ChartAxis{
			Font: excelize.Font{Size: 10},
		},
		YAxis: excelize.ChartAxis{
			Font: excelize.Font{Size: 10},
		},
		Dimension: excelize.ChartDimension{
			Width:  640,
			Height: 480,
		},
		Legend: excelize.ChartLegend{
			Position: "bottom",
		},
	}); err != nil {
		return err
	}
	
	return nil
}

// createEnhancedClusterSummarySheetWithCharts - создание листа сводки кластера.
func createEnhancedClusterSummarySheetWithCharts(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "📊 Сводка"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	title := fmt.Sprintf("📊 АНАЛИЗ РЕСУРСОВ KUBERNETES КЛАСТЕРА (запас %d%%)", bufferPercent)
	f.SetCellValue(sheetName, "A1", title)
	f.SetCellStyle(sheetName, "A1", "F1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "F1")
	f.SetRowHeight(sheetName, 1, 35)

	row := 3

	row = addSectionHeader(f, sheetName, row, "📦 ОБЩАЯ СТАТИСТИКА КЛАСТЕРА", styles)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего подов:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), cluster.TotalPods)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Всего нод:")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), cluster.TotalNodes)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles["good"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего неймспейсов:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(cluster.ByNamespace))
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Всего PVC:")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), cluster.TotalPVCs)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles["good"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего PV:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), cluster.TotalPVs)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	row += 2

	row = addSectionHeader(f, sheetName, row, "🔥 МАКСИМАЛЬНЫЕ ПОКАЗАТЕЛИ ПОДОВ", styles)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Макс. CPU (факт):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		fmt.Sprintf("%.0fm (%.2f ядер)", cluster.MaxPodCPUActual,
			cluster.MaxPodCPUActual/MillicoresInCore))
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), 
		fmt.Sprintf("Под: %s/%s", cluster.MaxPodNamespaceCPU, cluster.MaxPodNameCPU))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["warning"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Макс. память (факт):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		formatMemoryValue(cluster.MaxPodMemoryActual))
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row),
		fmt.Sprintf("Под: %s/%s", cluster.MaxPodNamespaceMemory, cluster.MaxPodNameMemory))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["warning"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("C%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Макс. CPU (request):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		fmt.Sprintf("%.0fm (%.2f ядер)", cluster.MaxPodCPURequest,
			cluster.MaxPodCPURequest/MillicoresInCore))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Макс. память (request):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		formatMemoryValue(cluster.MaxPodMemoryRequest))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row += 2

	row = addSectionHeader(f, sheetName, row, "⚙️  ПРОЦЕССОР (CPU)", styles)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Емкость кластера:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		fmt.Sprintf("%.0fm (%.1f ядер)", cluster.TotalNodeCPUCapacity,
			cluster.TotalNodeCPUCapacity/MillicoresInCore))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	row = addMetricRow(f, sheetName, row, "Запрошено (requests):",
		cluster.TotalCPURequest, cluster.TotalNodeCPUCapacity, true, styles)
	row = addMetricRow(f, sheetName, row, "Фактически (actual):",
		cluster.TotalCPUActual, cluster.TotalNodeCPUCapacity, true, styles)
	row = addMetricRow(f, sheetName, row, fmt.Sprintf("Оптимально (+%d%%):", bufferPercent),
		cluster.TotalCPUOptimized, cluster.TotalNodeCPUCapacity, true, styles)

	cpuDiff := cluster.TotalCPURequest - cluster.TotalCPUOptimized
	if cpuDiff > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "💰 Можно сэкономить:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
			fmt.Sprintf("%.0fm (%.2f ядер)", cpuDiff, cpuDiff/MillicoresInCore))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["good"])
	} else if cpuDiff < 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "⚠️ Требуется добавить:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
			fmt.Sprintf("%.0fm (%.2f ядер)", -cpuDiff, -cpuDiff/MillicoresInCore))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["warning"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["warning"])
	}
	row += 2

	row = addSectionHeader(f, sheetName, row, "💾 ПАМЯТЬ (MEMORY)", styles)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Емкость кластера:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		formatMemoryValue(cluster.TotalNodeMemoryCapacity))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	row = addMetricRow(f, sheetName, row, "Запрошено (requests):",
		cluster.TotalMemRequest, cluster.TotalNodeMemoryCapacity, false, styles)
	row = addMetricRow(f, sheetName, row, "Фактически (actual):",
		cluster.TotalMemActual, cluster.TotalNodeMemoryCapacity, false, styles)
	row = addMetricRow(f, sheetName, row, fmt.Sprintf("Оптимально (+%d%%):", bufferPercent),
		cluster.TotalMemOptimized, cluster.TotalNodeMemoryCapacity, false, styles)

	memDiff := cluster.TotalMemRequest - cluster.TotalMemOptimized
	if memDiff > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "💰 Можно сэкономить:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), formatMemoryValue(memDiff))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["good"])
	} else if memDiff < 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "⚠️ Требуется добавить:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), formatMemoryValue(-memDiff))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["warning"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["warning"])
	}
	row += 2

	row = addSectionHeader(f, sheetName, row, "📊 ЭФФЕКТИВНОСТЬ ИСПОЛЬЗОВАНИЯ", styles)

	cpuEff := (cluster.TotalCPUActual / cluster.TotalCPURequest) * 100
	memEff := (cluster.TotalMemActual / cluster.TotalMemRequest) * 100

	row = addEfficiencyRow(f, sheetName, row, "CPU эффективность:", cpuEff, styles)
	row = addEfficiencyRow(f, sheetName, row, "Память эффективность:", memEff, styles)
	row += 2

	row = addSectionHeader(f, sheetName, row, "💽 ХРАНИЛИЩА (STORAGE)", styles)
	
	pvcPercent := 0.0
	if cluster.TotalPVCCapacity > 0 {
		pvcPercent = (cluster.TotalPVCUsed / cluster.TotalPVCCapacity) * 100
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "PVC емкость:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		fmt.Sprintf("%s / %s (%.1f%%)",
			formatMemoryValue(cluster.TotalPVCUsed),
			formatMemoryValue(cluster.TotalPVCCapacity),
			pvcPercent))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row),
		getStyleByEfficiency(styles, pvcPercent))
	row++
	
	pvPercent := 0.0
	if cluster.TotalPVCapacity > 0 {
		pvPercent = (cluster.TotalPVUsed / cluster.TotalPVCapacity) * 100
	}
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "PV емкость:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
		fmt.Sprintf("%s / %s (%.1f%%)",
			formatMemoryValue(cluster.TotalPVUsed),
			formatMemoryValue(cluster.TotalPVCapacity),
			pvPercent))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row),
		getStyleByEfficiency(styles, pvPercent))

	f.SetColWidth(sheetName, "A", "A", 35)
	f.SetColWidth(sheetName, "B", "B", 50)
	f.SetColWidth(sheetName, "C", "F", 25)

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})
}

func boolPtr(b bool) *bool { return &b }

func addSectionHeader(f *excelize.File, sheetName string, row int, title string, styles map[string]int) int {
	cell := fmt.Sprintf("A%d", row)
	endCell := fmt.Sprintf("F%d", row)
	f.SetCellValue(sheetName, cell, title)
	f.MergeCell(sheetName, cell, endCell)
	f.SetCellStyle(sheetName, cell, endCell, styles["tableHeader"])
	f.SetRowHeight(sheetName, row, 25)
	return row + 1
}

func getStatus(eff float64) string {
	switch {
	case eff > EfficiencyCritical:
		return "Критично"
	case eff > EfficiencyHigh:
		return "Высокое"
	case eff > EfficiencyNormal:
		return "Норма"
	case eff > EfficiencyLow:
		return "Низкое"
	default:
		return "Минимум"
	}
}

func addMetricRow(f *excelize.File, sheetName string, row int, label string,
	value float64, total float64, isCPU bool, styles map[string]int) int {

	var valueStr string
	if isCPU {
		valueStr = fmt.Sprintf("%.0fm (%.2f ядер)", value, value/MillicoresInCore)
	} else {
		valueStr = formatMemoryValue(value)
	}

	percent := 0.0
	if total > 0 {
		percent = (value / total) * 100
	}

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), label)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), valueStr)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("%.1f%%", percent))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row),
		getStyleByEfficiency(styles, percent))
	return row + 1
}

func addEfficiencyRow(f *excelize.File, sheetName string, row int, label string,
	eff float64, styles map[string]int) int {

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), label)
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("%.1f%%", eff))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row),
		getStyleByEfficiency(styles, eff))
	return row + 1
}

// createGatekeeperSheet - создание листа с анализом Gatekeeper.
func createGatekeeperSheet(f *excelize.File, cluster *ClusterSummary, styles map[string]int) {
	sheetName := "🔒 Gatekeeper"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "🔒 АНАЛИЗ ПОЛИТИК OPA GATEKEEPER")
	f.SetCellStyle(sheetName, "A1", "G1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "G1")
	f.SetRowHeight(sheetName, 1, 35)

	row := 3
	gk := cluster.Gatekeeper

	row = addSectionHeader(f, sheetName, row, "📋 СТАТУС GATEKEEPER", styles)

	statusText := "Не установлен"
	statusStyle := styles["warning"]
	if gk != nil && gk.Installed {
		if gk.Running {
			statusText = fmt.Sprintf("Активен (%d под(ов) работает)", gk.PodCount)
			statusStyle = styles["good"]
		} else {
			statusText = "Установлен, но не запущен"
			statusStyle = styles["critical"]
		}
	}

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Статус:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), statusText)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), statusStyle)
	row++

	if gk == nil || !gk.Installed {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Рекомендация:")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row),
			"Gatekeeper не обнаружен. Рекомендуется установить OPA Gatekeeper для управления политиками кластера.")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["warning"])
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("G%d", row), styles["data"])
		f.MergeCell(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("G%d", row))

		f.SetColWidth(sheetName, "A", "A", 30)
		f.SetColWidth(sheetName, "B", "G", 25)
		return
	}

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Шаблонов ограничений:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(gk.ConstraintTemplates))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Активных ограничений:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(gk.Constraints))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row += 2

	if len(gk.ConstraintTemplates) > 0 {
		row = addSectionHeader(f, sheetName, row, "📄 ШАБЛОНЫ ОГРАНИЧЕНИЙ (ConstraintTemplates)", styles)

		templateHeaders := []string{"Имя шаблона", "Kind ограничения"}
		for i, h := range templateHeaders {
			cell := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue(sheetName, cell, h)
			f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
		}
		f.SetRowHeight(sheetName, row, 25)
		row++

		for _, tmpl := range gk.ConstraintTemplates {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), tmpl.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), tmpl.Kind)
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["data"])
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
			row++
		}
		row++
	}

	if len(gk.Constraints) > 0 {
		row = addSectionHeader(f, sheetName, row, "⚙️  АКТИВНЫЕ ОГРАНИЧЕНИЯ (Constraints)", styles)

		constraintHeaders := []string{"Имя", "Тип", "Режим", "Нарушений", "Неймспейсы"}
		for i, h := range constraintHeaders {
			cell := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue(sheetName, cell, h)
			f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
		}
		f.SetRowHeight(sheetName, row, 25)
		row++

		for _, c := range gk.Constraints {
			enfStyle := styles["good"]
			switch c.EnforcementAction {
			case "warn":
				enfStyle = styles["warning"]
			case "dryrun":
				enfStyle = styles["low"]
			}

			violStyle := styles["data"]
			if c.TotalViolations > 0 {
				violStyle = styles["critical"]
			}

			nsText := "Все неймспейсы"
			if len(c.Namespaces) > 0 {
				nsText = strings.Join(c.Namespaces, ", ")
			}

			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), c.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), c.Kind)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), c.EnforcementAction)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), c.TotalViolations)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), nsText)
			f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["data"])
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
			f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), enfStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), violStyle)
			f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles["data"])
			row++
		}

		f.AutoFilter(sheetName, fmt.Sprintf("A%d:E%d", row-len(gk.Constraints)-1, row-len(gk.Constraints)-1), nil)
	}

	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 50)

	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})
}

// createHistorySheet - создание листа с историческими метриками (min/avg/max/p95).
func createHistorySheet(f *excelize.File, histories map[string]*PodHistory, styles map[string]int) {
	sheetName := "📈 История"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "📈 ИСТОРИЧЕСКИЕ МЕТРИКИ ПОДОВ")
	f.SetCellStyle(sheetName, "A1", "L1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "L1")
	f.SetRowHeight(sheetName, 1, 35)

	modeLabel := "Живой сбор"
	if prometheusURL != "" {
		modeLabel = fmt.Sprintf("Prometheus: %s", prometheusURL)
	}
	periodLabel := collectDuration
	if periodLabel == "" {
		periodLabel = "разовый"
	}
	f.SetCellValue(sheetName, "A2",
		fmt.Sprintf("Режим: %s  |  Период: %s", modeLabel, periodLabel))
	f.SetCellStyle(sheetName, "A2", "L2", styles["data"])
	f.MergeCell(sheetName, "A2", "L2")
	f.SetRowHeight(sheetName, 2, 20)

	row := 4

	// Заголовки таблицы
	headers := []string{
		"Неймспейс", "Под",
		"CPU Min", "CPU Avg", "CPU Max", "CPU P95",
		"Mem Min", "Mem Avg", "Mem Max", "Mem P95",
		"Семплов", "Статус CPU",
	}
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 30)
	headerRow := row
	row++

	// Сортируем поды по неймспейсу и имени
	type histEntry struct {
		key  string
		hist *PodHistory
	}
	var entries []histEntry
	for k, h := range histories {
		entries = append(entries, histEntry{k, h})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].hist.Namespace != entries[j].hist.Namespace {
			return entries[i].hist.Namespace < entries[j].hist.Namespace
		}
		return entries[i].hist.Name < entries[j].hist.Name
	})

	for _, e := range entries {
		h := e.hist
		if h.SampleCount == 0 {
			continue
		}

		// Определяем стиль по P95 CPU относительно среднего (признак нестабильности)
		cpuVolatility := 0.0
		if h.CPUAvg > 0 {
			cpuVolatility = ((h.CPUP95 - h.CPUAvg) / h.CPUAvg) * 100
		}
		cpuStyle := styles["data"]
		statusLabel := "Стабильно"
		switch {
		case cpuVolatility > 100:
			cpuStyle = styles["critical"]
			statusLabel = "Очень нестабильно"
		case cpuVolatility > 50:
			cpuStyle = styles["high"]
			statusLabel = "Нестабильно"
		case cpuVolatility > 20:
			cpuStyle = styles["low"]
			statusLabel = "Небольшие пики"
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), h.Namespace)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), h.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), formatCPUValue(h.CPUMin))
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), formatCPUValue(h.CPUAvg))
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), formatCPUValue(h.CPUMax))
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), formatCPUValue(h.CPUP95))
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), formatMemoryValue(h.MemMin))
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), formatMemoryValue(h.MemAvg))
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), formatMemoryValue(h.MemMax))
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), formatMemoryValue(h.MemP95))
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), h.SampleCount)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), statusLabel)

		for _, col := range []string{"A", "B", "C", "D", "E", "G", "H", "I", "J", "K"} {
			f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, row), fmt.Sprintf("%s%d", col, row), styles["data"])
		}
		// P95 CPU и статус — цветные
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), cpuStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), cpuStyle)

		row++
	}

	// Автофильтр
	f.AutoFilter(sheetName, fmt.Sprintf("A%d:L%d", headerRow, headerRow), nil)

	// Ширина колонок
	colWidths := map[string]float64{
		"A": 25, "B": 45, "C": 14, "D": 14, "E": 14, "F": 14,
		"G": 14, "H": 14, "I": 14, "J": 14, "K": 12, "L": 20,
	}
	for col, w := range colWidths {
		f.SetColWidth(sheetName, col, col, w)
	}

	// Закрепляем заголовки
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      2,
		YSplit:      headerRow,
		TopLeftCell: fmt.Sprintf("C%d", headerRow+1),
		ActivePane:  "bottomRight",
	})
}

// createRBACSheet - создание листа с анализом прав доступа
func createRBACSheet(f *excelize.File, cluster *ClusterSummary, styles map[string]int) {
	sheetName := "👥 RBAC"
	f.NewSheet(sheetName)

	f.SetCellValue(sheetName, "A1", "👥 ПРАВА ДОСТУПА КЛАСТЕРА (RBAC)")
	f.SetCellStyle(sheetName, "A1", "H1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "H1")
	f.SetRowHeight(sheetName, 1, 35)

	row := 3

	// Статистика
	row = addSectionHeader(f, sheetName, row, "📊 СТАТИСТИКА RBAC", styles)

	clusterLevel := 0
	nsLevel := 0
	users := make(map[string]bool)
	sas := make(map[string]bool)
	groups := make(map[string]bool)

	for _, e := range cluster.RBACEntries {
		if e.Scope == "cluster" {
			clusterLevel++
		} else {
			nsLevel++
		}
		switch e.SubjectKind {
		case "User":
			users[e.Subject] = true
		case "ServiceAccount":
			sas[e.Subject+"/"+e.SubjectNS] = true
		case "Group":
			groups[e.Subject] = true
		}
	}

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Привязок на уровне кластера:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), clusterLevel)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Привязок на уровне неймспейса:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), nsLevel)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Уникальных пользователей (User):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(users))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Сервисных аккаунтов (ServiceAccount):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(sas))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row++

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Групп (Group):")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(groups))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
	row += 2

	// Таблица привязок ролей
	row = addSectionHeader(f, sheetName, row, "🔑 ПОЛНЫЙ СПИСОК ПРИВЯЗОК РОЛЕЙ", styles)

	headers := []string{
		"Субъект", "Тип субъекта", "Неймспейс субъекта",
		"Роль", "Тип роли", "Привязка", "Тип привязки", "Область / Неймспейс",
	}
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 30)
	headerRow := row
	row++

	// Сортируем записи: сначала Users, затем Groups, затем ServiceAccounts
	sorted := make([]*RBACEntry, len(cluster.RBACEntries))
	copy(sorted, cluster.RBACEntries)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].SubjectKind != sorted[j].SubjectKind {
			order := map[string]int{"User": 0, "Group": 1, "ServiceAccount": 2}
			return order[sorted[i].SubjectKind] < order[sorted[j].SubjectKind]
		}
		if sorted[i].Subject != sorted[j].Subject {
			return sorted[i].Subject < sorted[j].Subject
		}
		return sorted[i].BoundIn < sorted[j].BoundIn
	})

	for _, e := range sorted {
		subjectStyle := styles["data"]
		switch e.SubjectKind {
		case "User":
			subjectStyle = styles["good"]
		case "Group":
			subjectStyle = styles["warning"]
		}

		scopeText := "Весь кластер"
		if e.BoundIn != "" {
			scopeText = e.BoundIn
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), e.Subject)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), e.SubjectKind)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), e.SubjectNS)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), e.Role)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), e.RoleKind)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), e.BindingName)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), e.BindingKind)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), scopeText)

		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), subjectStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), styles["data"])
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), styles["data"])
		row++
	}

	// Автофильтр
	lastDataCol := "H"
	f.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastDataCol, headerRow), nil)

	// Ширина колонок
	colWidths := []float64{35, 20, 25, 35, 15, 45, 20, 30}
	for i, width := range colWidths {
		f.SetColWidth(sheetName, string(rune('A'+i)), string(rune('A'+i)), width)
	}

	// Закрепляем панели
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      1,
		YSplit:      headerRow,
		TopLeftCell: fmt.Sprintf("B%d", headerRow+1),
		ActivePane:  "bottomRight",
	})
}