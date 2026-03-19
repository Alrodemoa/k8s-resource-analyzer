package main

// Модуль генерации Excel отчётов
// Все функции для создания детальных Excel файлов с анализом кластера

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

// ============================================================================
// EXCEL REPORT GENERATION
// ============================================================================

// generateExcelReport - генерация Excel отчета
func generateExcelReport(cluster *ClusterSummary) string {
	printStep("📁 Создаем Excel отчет...")

	f := excelize.NewFile()
	defer f.Close()

	// Создаем стили
	styles := createEnhancedStyles(f)

	// Создаем листы (первый созданный лист станет активным)
	createEnhancedClusterSummarySheetWithCharts(f, cluster, styles)
	createEnhancedNodesSheet(f, cluster, styles)
	createEnhancedPVCSheet(f, cluster, styles)
	createEnhancedPVSheet(f, cluster, styles)
	
	// Создаём листы неймспейсов с полными данными по подам
	// Нужно собрать данные по подам заново
	allPods := collectAllPodsForReport(cluster)
	createEnhancedNamespaceSheetsWithPods(f, cluster, allPods, styles)

	// Удаляем дефолтный лист Sheet1 (он создаётся автоматически)
	if index, err := f.GetSheetIndex("Sheet1"); err == nil && index >= 0 {
		f.DeleteSheet("Sheet1")
	}

	// Устанавливаем первый лист активным
	f.SetActiveSheet(0)

	// Сохраняем файл
	filename := fmt.Sprintf("k8s-анализ-%s.xlsx",
		time.Now().Format("2006-01-02-150405"))
	if err := f.SaveAs(filename); err != nil {
		printError(fmt.Sprintf("❌ Ошибка сохранения: %v", err))
		os.Exit(1)
	}

	printSuccess("   ✅ Excel отчет создан")
	return filename
}

// createEnhancedStyles - создание улучшенных стилей для Excel с приятной цветовой гаммой
func createEnhancedStyles(f *excelize.File) map[string]int {
	styles := make(map[string]int)

	// Заголовок главный (крупный, элегантный тёмно-фиолетовый)
	mainHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#5B4A8F"}, // Элегантный фиолетовый
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

	// Заголовок таблицы (средний, приятный индиго)
	tableHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#6B5B95"}, // Мягкий индиго
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

	// Критический (мягкий коралловый красный)
	critical, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E74C3C"}, // Мягкий коралловый красный
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

	// Высокий (тёплый янтарный)
	high, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#000000",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F39C12"}, // Янтарный оранжевый
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

	// Норма (свежий мятный зелёный)
	normal, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#27AE60"}, // Свежий зелёный
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

	// Низкий (мягкий золотистый)
	low, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#000000",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F4D03F"}, // Мягкий золотой
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

	// Данные (чистый белый фон с переносом строк)
	data, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
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

	// Числа
	number, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
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

	// Хороший показатель (нежный мятный)
	good, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#1E8449",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D5F4E6"}, // Нежный мятный
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

	// Предупреждение (персиковый)
	warning, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Color:  "#9C5700",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FCE4C8"}, // Нежный персиковый
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


// Вспомогательные функции для форматирования Excel
func createEnhancedNodesSheet(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "🖥️ Ноды"
	f.NewSheet(sheetName)

	// Заголовок
	f.SetCellValue(sheetName, "A1", "🖥️ АНАЛИЗ ЗАГРУЗКИ НОД КЛАСТЕРА")
	f.SetCellStyle(sheetName, "A1", "M1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "M1")
	f.SetRowHeight(sheetName, 1, 30)

	// Заголовки колонок
	headers := []string{
		"Нода", "Поды",
		"CPU\nЕмкость", "CPU\nRequests", "CPU\nActual", "CPU Req\n%", "CPU Act\n%",
		"Память\nЕмкость", "Память\nRequests", "Память\nActual", "Mem Req\n%", "Mem Act\n%",
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

	// Данные нод
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

		// CPU метрики
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

		// Memory метрики
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

		// Рекомендации
		f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+12, row), node.Recommendation)
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+12, row),
			fmt.Sprintf("%c%d", col+12, row), styles["data"])

		row++
	}

	// Ширина колонок
	colWidths := []float64{30, 10, 15, 15, 15, 12, 12, 18, 18, 18, 12, 12, 50}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	// Автофильтр
	lastCol := string(rune('A' + len(headers) - 1))
	f.AutoFilter(sheetName, fmt.Sprintf("A3:%s3", lastCol), nil)

	// Закрепляем панели
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
		"Емкость", "Запрошено", "StorageClass", "Режимы доступа",
	}

	row := 3
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 25)
	row++

	// Сортируем PVC
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

		// Стиль для статуса
		statusStyle := styles["good"]
		if pvc.Status != "Bound" {
			statusStyle = styles["warning"]
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
			fmt.Sprintf("%c%d", col+2, row), statusStyle)

		row++
	}

	// Ширина колонок
	colWidths := []float64{25, 35, 15, 35, 15, 15, 25, 30}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	// Автофильтр
	f.AutoFilter(sheetName, "A3:H3", nil)

	// Закрепляем панели
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
		"Имя PV", "Емкость", "Статус", "Claim", "StorageClass",
	}

	row := 3
	for i, h := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
	}
	f.SetRowHeight(sheetName, row, 25)
	row++

	// Сортируем PV
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

		// Стиль для статуса
		statusStyle := styles["good"]
		if pv.Status != "Bound" {
			statusStyle = styles["warning"]
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
			fmt.Sprintf("%c%d", col+2, row), statusStyle)

		row++
	}

	// Ширина колонок
	colWidths := []float64{40, 15, 15, 45, 25}
	for i, width := range colWidths {
		colName := string(rune('A' + i))
		f.SetColWidth(sheetName, colName, colName, width)
	}

	// Автофильтр
	f.AutoFilter(sheetName, "A3:E3", nil)

	// Закрепляем панели
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      3,
		TopLeftCell: "A4",
		ActivePane:  "bottomLeft",
	})
}


// ============================================================================
// USER INTERACTION
// ============================================================================

// interactiveNamespaceSelect - интерактивный выбор неймспейсов
func collectAllPodsForReport(cluster *ClusterSummary) map[string]map[string]*PodResource {
	allPods := make(map[string]map[string]*PodResource)
	
	// Собираем поды по каждому неймспейсу из кластера
	for ns := range cluster.ByNamespace {
		// Получаем requests и limits
		pods := getPodResources(ns)
		if len(pods) > 0 {
			if allPods[ns] == nil {
				allPods[ns] = make(map[string]*PodResource)
			}
			for _, pod := range pods {
				allPods[ns][pod.Name] = pod
			}
		}
		
		// Получаем actual usage
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

// createEnhancedNamespaceSheetsWithPods - создание детальных листов неймспейсов с подами
func createEnhancedNamespaceSheetsWithPods(f *excelize.File, cluster *ClusterSummary,
	allPods map[string]map[string]*PodResource, styles map[string]int) {
	
	printStep("   Создаем детальные листы неймспейсов...")
	
	// Сортируем неймспейсы для удобства
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
		
		// ===== ШАПКА НЕЙМСПЕЙСА =====
		title := fmt.Sprintf("📊 НЕЙМСПЕЙС: %s", ns)
		f.SetCellValue(sheetName, "A1", title)
		f.SetCellStyle(sheetName, "A1", "N1", styles["mainHeader"])
		f.MergeCell(sheetName, "A1", "N1")
		f.SetRowHeight(sheetName, 1, 35)
		
		row := 3
		
		// ===== ДЕТАЛЬНАЯ СТАТИСТИКА НЕЙМСПЕЙСА =====
		row = addSectionHeader(f, sheetName, row, "📊 СТАТИСТИКА НЕЙМСПЕЙСА", styles)
		
		// Первая строка статистики
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
		
		// Вторая строка статистики
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
		
		// Третья строка статистики
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
		
		// ===== ТАБЛИЦА С ПОДАМИ =====
		row = addSectionHeader(f, sheetName, row, "📦 ДЕТАЛЬНАЯ ИНФОРМАЦИЯ ПО ПОДАМ", styles)
		
		// Заголовки таблицы
		headers := []string{
			"Имя пода", "Нода", "PVC",
			"CPU\nRequest", "CPU\nActual", "CPU\nLimit", "CPU\nЭфф %",
			"Память\nRequest", "Память\nActual", "Память\nLimit", "Память\nЭфф %",
			"Рекомендации", "Статус",
		}
		
		for i, h := range headers {
			cell := fmt.Sprintf("%c%d", 'A'+i, row)
			f.SetCellValue(sheetName, cell, h)
			f.SetCellStyle(sheetName, cell, cell, styles["tableHeader"])
		}
		f.SetRowHeight(sheetName, row, 35)
		row++
		
		// Данные подов
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
			
			// Имя пода
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col, row), pod.Name)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col, row), 
				fmt.Sprintf("%c%d", col, row), styles["data"])
			
			// Нода
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+1, row), pod.NodeName)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+1, row),
				fmt.Sprintf("%c%d", col+1, row), styles["data"])
			
			// PVC
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+2, row), 
				strings.Join(pod.PVCs, ", "))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+2, row),
				fmt.Sprintf("%c%d", col+2, row), styles["data"])
			
			// CPU Request
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+3, row), pod.CPURequest)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+3, row),
				fmt.Sprintf("%c%d", col+3, row), styles["number"])
			
			// CPU Actual
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+4, row), pod.CPUActual)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+4, row),
				fmt.Sprintf("%c%d", col+4, row), styles["number"])
			
			// CPU Limit
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+5, row), pod.CPULimit)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+5, row),
				fmt.Sprintf("%c%d", col+5, row), styles["number"])
			
			// CPU Эффективность
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+6, row), 
				fmt.Sprintf("%.1f%%", cpuEff))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+6, row),
				fmt.Sprintf("%c%d", col+6, row), getStyleByEfficiency(styles, cpuEff))
			
			// Memory Request
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+7, row), pod.MemoryRequest)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+7, row),
				fmt.Sprintf("%c%d", col+7, row), styles["number"])
			
			// Memory Actual
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+8, row), pod.MemoryActual)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+8, row),
				fmt.Sprintf("%c%d", col+8, row), styles["number"])
			
			// Memory Limit
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+9, row), pod.MemoryLimit)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+9, row),
				fmt.Sprintf("%c%d", col+9, row), styles["number"])
			
			// Memory Эффективность
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+10, row), 
				fmt.Sprintf("%.1f%%", memEff))
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+10, row),
				fmt.Sprintf("%c%d", col+10, row), getStyleByEfficiency(styles, memEff))
			
			// Рекомендации
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+11, row), recommendation)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+11, row),
				fmt.Sprintf("%c%d", col+11, row), styles["data"])
			
			// Статус
			f.SetCellValue(sheetName, fmt.Sprintf("%c%d", col+12, row), status)
			f.SetCellStyle(sheetName, fmt.Sprintf("%c%d", col+12, row),
				fmt.Sprintf("%c%d", col+12, row), getStyleByEfficiency(styles, memEff))
			
			row++
		}
		
		// Ширина колонок
		colWidths := []float64{40, 25, 30, 15, 15, 15, 12, 18, 18, 18, 12, 80, 18}
		for i, width := range colWidths {
			colName := string(rune('A' + i))
			f.SetColWidth(sheetName, colName, colName, width)
		}
		
		// Автофильтр
		lastCol := string(rune('A' + len(headers) - 1))
		headerRow := row - len(podList) - 1
		f.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastCol, headerRow), nil)
		
		// Закрепляем панели
		f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			XSplit:      1,
			YSplit:      headerRow,
			TopLeftCell: fmt.Sprintf("B%d", headerRow+1),
			ActivePane:  "bottomRight",
		})
	}
}

// addChartToSheet - добавление графика на лист
func addChartToSheet(f *excelize.File, sheetName string, 
	chartType string, title string, 
	categories string, values string, 
	position string) error {
	
	// Создаём график
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
			PrintObject:     excelize.Bool(true),
			LockAspectRatio: false,
			Locked:          excelize.Bool(false),
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

// createEnhancedClusterSummarySheetWithCharts - создание расширенного листа сводки с графиками
func createEnhancedClusterSummarySheetWithCharts(f *excelize.File, cluster *ClusterSummary,
	styles map[string]int) {

	sheetName := "📊 Сводка"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	// Заголовок
	title := fmt.Sprintf("📊 АНАЛИЗ РЕСУРСОВ KUBERNETES КЛАСТЕРА (запас %d%%)", bufferPercent)
	f.SetCellValue(sheetName, "A1", title)
	f.SetCellStyle(sheetName, "A1", "F1", styles["mainHeader"])
	f.MergeCell(sheetName, "A1", "F1")
	f.SetRowHeight(sheetName, 1, 35)

	row := 3

	// ===== СЕКЦИЯ 1: ОБЩАЯ СТАТИСТИКА =====
	row = addSectionHeader(f, sheetName, row, "📦 ОБЩАЯ СТАТИСТИКА КЛАСТЕРА", styles)
	
	// Строка 1
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего подов:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), cluster.TotalPods)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Всего нод:")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), cluster.TotalNodes)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles["good"])
	row++
	
	// Строка 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего неймспейсов:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), len(cluster.ByNamespace))
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), "Всего PVC:")
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), cluster.TotalPVCs)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), styles["good"])
	row++
	
	// Строка 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Всего PV:")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), cluster.TotalPVs)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), styles["good"])
	row += 2

	// ===== СЕКЦИЯ 2: МАКСИМАЛЬНЫЕ ПОКАЗАТЕЛИ =====
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

	// ===== СЕКЦИЯ 3: ПРОЦЕССОР (CPU) =====
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
	
	// Добавляем экономию/дефицит
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

	// ===== СЕКЦИЯ 4: ПАМЯТЬ (MEMORY) =====
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
	
	// Добавляем экономию/дефицит
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

	// ===== СЕКЦИЯ 5: ЭФФЕКТИВНОСТЬ =====
	row = addSectionHeader(f, sheetName, row, "📊 ЭФФЕКТИВНОСТЬ ИСПОЛЬЗОВАНИЯ", styles)
	
	cpuEff := (cluster.TotalCPUActual / cluster.TotalCPURequest) * 100
	memEff := (cluster.TotalMemActual / cluster.TotalMemRequest) * 100
	
	row = addEfficiencyRow(f, sheetName, row, "CPU эффективность:", cpuEff, styles)
	row = addEfficiencyRow(f, sheetName, row, "Память эффективность:", memEff, styles)
	row += 2
	
	// ===== СЕКЦИЯ 6: ХРАНИЛИЩА =====
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

	// Ширина колонок
	f.SetColWidth(sheetName, "A", "A", 35)
	f.SetColWidth(sheetName, "B", "B", 50)
	f.SetColWidth(sheetName, "C", "F", 25)

	// Закрепляем панели
	f.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      0,
		YSplit:      2,
		TopLeftCell: "A3",
		ActivePane:  "bottomLeft",
	})
	
	// Примечание: Графики будут добавляться в отдельных листах для лучшей визуализации
	// так как в сводке уже много данных
}

// sanitizeSheetName - очистка имени листа Excel