// K8s Resource Analyzer
//
// Анализатор ресурсов Kubernetes кластера с Excel отчётностью
//
// МОДУЛИ ПРОЕКТА:
// - main.go       - Точка входа программы
// - constants.go  - Константы и глобальные переменные
// - types.go      - Структуры данных
// - kubernetes.go - Работа с kubectl API
// - utils.go      - Утилиты (парсеры, форматтеры)
// - collector.go  - Сбор данных из кластера
// - analyzer.go   - Анализ и расчёты
// - excel.go      - Генерация Excel отчёта
// - ui.go         - Консольный интерфейс

package main

import (
	"flag"
	"os"
	"time"
)

func main() {
	// Парсим аргументы командной строки
	parseFlags()

	startTime := time.Now()
	printBanner()

	// Проверяем окружение
	if !validateEnvironment() {
		return
	}

	// Собираем данные кластера
	clusterSummary := collectClusterData()

	// Создаем Excel отчет
	filename := generateExcelReport(clusterSummary)

	// Выводим итоговую статистику
	printFinalSummary(clusterSummary, startTime, filename)
}

// parseFlags - парсинг флагов командной строки
func parseFlags() {
	var help bool
	var version bool
	var buffer int

	flag.IntVar(&buffer, "buffer", 50, "Процент запаса ресурсов (по умолчанию 50)")
	flag.IntVar(&buffer, "b", 50, "Процент запаса ресурсов (сокращенно)")
	flag.BoolVar(&help, "help", false, "Показать справку")
	flag.BoolVar(&help, "h", false, "Показать справку (сокращенно)")
	flag.BoolVar(&version, "version", false, "Показать версию программы")
	flag.BoolVar(&version, "v", false, "Показать версию программы (сокращенно)")
	flag.Parse()

	bufferPercent = buffer

	// Версия выводится первой и завершает программу
	if version {
		printVersion()
		os.Exit(0)
	}

	if help {
		printHelp()
		os.Exit(0)
	}
}

// validateEnvironment - проверка окружения и инициализация клиента
func validateEnvironment() bool {
	printStep("🔌 Подключение к Kubernetes API...")

	if !checkKubernetesConnection() {
		printError("❌ Не удалось подключиться к кластеру")
		printError("💡 Проверьте:")
		printError("   • Переменная KUBECONFIG указывает на правильный файл")
		printError("   • Файл ~/.kube/config существует и корректен")
		printError("   • У вас есть доступ к кластеру")
		return false
	}

	printStep("✅ Подключение установлено")
	return true
}
