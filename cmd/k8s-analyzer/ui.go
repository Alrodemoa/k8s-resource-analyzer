package main

// Модуль консольного интерфейса

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ANSI коды цветов и стилей
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiCyan   = "\033[96m"
	ansiYellow = "\033[93m"
	ansiGreen  = "\033[92m"
	ansiBlue   = "\033[94m"
	ansiRed    = "\033[91m"
	ansiWhite  = "\033[97m"
	ansiGray   = "\033[90m"
)

// printBanner - вывод баннера программы
func printBanner() {
	sep := ansiCyan + strings.Repeat("─", ConsoleWidth) + ansiReset

	// ASCII-арт "K8s" + информация справа
	logo := []string{
		ansiYellow + ansiBold + `  ██╗  ██╗  █████╗   ██████╗  ` + ansiReset + `  ` + ansiWhite + ansiBold + `Анализатор ресурсов` + ansiReset,
		ansiYellow + ansiBold + `  ██║ ██╔╝ ██╔══██╗ ██╔════╝  ` + ansiReset + `  ` + ansiGray + strings.Repeat("─", 40) + ansiReset,
		ansiYellow + ansiBold + `  █████╔╝  ╚█████╔╝ ╚█████╗   ` + ansiReset + `  ` + ansiGreen + ansiBold + `v` + AppVersion + ansiReset + ansiGray + `  ·  Kubernetes Кластер` + ansiReset,
		ansiYellow + ansiBold + `  ██╔═██╗  ██╔══██╗  ╚════██╗ ` + ansiReset + `  ` + ansiBlue + `Анализ и оптимизация ресурсов` + ansiReset,
		ansiYellow + ansiBold + `  ██║  ██╗ ╚█████╔╝  ██████╔╝ ` + ansiReset + `  ` + ansiBlue + `кластера с Excel отчётом` + ansiReset,
		ansiYellow + ansiBold + `  ╚═╝  ╚═╝  ╚════╝   ╚═════╝  ` + ansiReset,
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println()
	for _, line := range logo {
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Println(sep)
	fmt.Println()
	fmt.Printf(ansiGray+"  📅 Дата:           "+ansiReset+ansiWhite+"%s"+ansiReset+"\n",
		time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf(ansiGray+"  🔧 Запас ресурсов: "+ansiReset+ansiWhite+"%d%%"+ansiReset+"\n",
		bufferPercent)

	// Режим работы
	if prometheusURL != "" && collectDuration != "" {
		fmt.Printf(ansiGray+"  📡 Режим:          "+ansiReset+ansiCyan+"Prometheus (%s, период: %s)"+ansiReset+"\n",
			prometheusURL, collectDuration)
	} else if collectDuration != "" {
		fmt.Printf(ansiGray+"  ⏱️  Режим:          "+ansiReset+ansiCyan+"Живой сбор (период: %s)"+ansiReset+"\n",
			collectDuration)
	} else {
		fmt.Println(ansiGray + "  📊 Режим:          " + ansiReset + ansiWhite + "Текущий момент" + ansiReset)
	}
	fmt.Println()
}

// printVersion - вывод версии программы
func printVersion() {
	fmt.Printf("k8s-analyzer v%s\n", AppVersion)
}

// printHelp - вывод справки
func printHelp() {
	sep := ansiCyan + strings.Repeat("─", ConsoleWidth) + ansiReset
	fmt.Println()
	fmt.Println(sep)
	fmt.Println(ansiWhite + ansiBold + "  Использование:" + ansiReset)
	fmt.Println(sep)
	fmt.Println()
	fmt.Println(ansiWhite + "  k8s-analyzer" + ansiReset + ansiGray + " [опции]" + ansiReset)
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Опции:" + ansiReset)
	fmt.Println(ansiGreen + "    -b, --buffer <число>" + ansiReset + "                    Процент запаса ресурсов " + ansiGray + "(по умолчанию 50)" + ansiReset)
	fmt.Println(ansiGreen + "    -p, --prometheus <url>" + ansiReset + "                  URL Prometheus/Thanos " + ansiGray + "(например: http://prometheus:9090)" + ansiReset)
	fmt.Println(ansiGreen + "    -d, --duration <период>" + ansiReset + "                 Период анализа " + ansiGray + "(например: 30m, 2h, 7d, 1w)" + ansiReset)
	fmt.Println(ansiGreen + "    --cluster <имя>" + ansiReset + "                         Имя кластера в Thanos " + ansiGray + "(автоопределение если не указано)" + ansiReset)
	fmt.Println(ansiGreen + "    --cluster-label <лейбл>" + ansiReset + "                Лейбл кластера в Thanos " + ansiGray + "(автоопределение, обычно 'cluster')" + ansiReset)
	fmt.Println(ansiGreen + "  -k, --insecure" + ansiReset + "                            Не проверять TLS-сертификат " + ansiGray + "(как -k в curl/kubectl)" + ansiReset)
	fmt.Println(ansiGreen + "    -v, --version" + ansiReset + "                           Показать версию программы")
	fmt.Println(ansiGreen + "    -h, --help" + ansiReset + "                              Показать эту справку")
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Примеры:" + ansiReset)
	fmt.Println(ansiGray + "    k8s-analyzer" + ansiReset + "                             # Текущий момент, запас 50%")
	fmt.Println(ansiGray + "    k8s-analyzer -b 30" + ansiReset + "                       # Запас 30% (dev-окружение)")
	fmt.Println(ansiGray + "    k8s-analyzer -d 10m" + ansiReset + "                      # Живой сбор 10 минут")
	fmt.Println(ansiGray + "    k8s-analyzer -p http://prometheus:9090 -d 1h" + ansiReset + " # История из Prometheus за 1 час")
	fmt.Println(ansiGray + "    k8s-analyzer -p http://thanos:9090 -d 30m" + ansiReset + "  # История из Thanos за 30 минут")
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Требования:" + ansiReset)
	fmt.Println("    • Доступ к Kubernetes кластеру " + ansiGray + "(kubeconfig)" + ansiReset)
	fmt.Println("    • Metrics Server установлен в кластере")
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Конфигурация:" + ansiReset)
	fmt.Println("    • " + ansiGray + "~/.kube/config" + ansiReset + " (по умолчанию)")
	fmt.Println("    • Переменная окружения " + ansiGray + "KUBECONFIG" + ansiReset)
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Результат:" + ansiReset)
	fmt.Println("    " + ansiGreen + "k8s-анализ-YYYY-MM-DD-HHMMSS.xlsx" + ansiReset)
	fmt.Println()
	fmt.Println(sep)
	fmt.Println()
}

// printStep - вывод шага выполнения
func printStep(message string) {
	fmt.Println(message)
}

// printError - вывод ошибки
func printError(message string) {
	fmt.Println(ansiRed + message + ansiReset)
}

// printSuccess - вывод успешного сообщения
func printSuccess(message string) {
	fmt.Println(ansiGreen + message + ansiReset)
}

// printFinalSummary - вывод итоговой статистики
func printFinalSummary(cluster *ClusterSummary, startTime time.Time, filename string) {
	elapsed := time.Since(startTime)
	sep := ansiCyan + strings.Repeat("═", ConsoleWidth) + ansiReset
	thin := ansiGray + strings.Repeat("─", ConsoleWidth) + ansiReset

	fmt.Println("\n" + sep)
	fmt.Println(centerText(ansiWhite+ansiBold+"  ИТОГОВАЯ СТАТИСТИКА  "+ansiReset, ConsoleWidth+len(ansiWhite+ansiBold+ansiReset)))
	fmt.Println(sep)

	fmt.Println()
	fmt.Printf(ansiCyan+"  🖥️  Нод:"+ansiReset+"          "+ansiWhite+"%d"+ansiReset+"\n", cluster.TotalNodes)
	fmt.Printf(ansiCyan+"  📦 Подов:"+ansiReset+"         "+ansiWhite+"%d"+ansiReset+"\n", cluster.TotalPods)
	fmt.Printf(ansiCyan+"  📂 Неймспейсов:"+ansiReset+"   "+ansiWhite+"%d"+ansiReset+"\n", len(cluster.ByNamespace))
	fmt.Printf(ansiCyan+"  💽 PVC / PV:"+ansiReset+"      "+ansiWhite+"%d / %d"+ansiReset+"\n", cluster.TotalPVCs, cluster.TotalPVs)

	fmt.Println()
	fmt.Println(thin)

	bufMul := 1.0 + float64(bufferPercent)/100.0

	// CPU
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  💻 CPU" + ansiReset)
	fmt.Printf(ansiGray+"     Фактически:    "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPUActual))
	fmt.Printf(ansiGray+"     Requests:      "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPURequest))
	if cluster.TotalCPULimit > 0 {
		cpuLimRatio := (cluster.TotalCPUActual / cluster.TotalCPULimit) * 100.0
		fmt.Printf(ansiGray+"     Limits:        "+ansiReset+ansiWhite+"%-12s"+ansiReset+ansiGray+"  факт: %.0f%% от limits"+ansiReset+"\n",
			formatCPUValue(cluster.TotalCPULimit), cpuLimRatio)
		// Потенциальная оптимизация лимитов с учётом буфера
		optimalLim := cluster.TotalCPUActual * bufMul
		if cluster.TotalCPULimit > optimalLim*1.1 {
			savings := cluster.TotalCPULimit - optimalLim
			fmt.Printf(ansiGreen+"     💡 Limits можно снизить до: "+ansiReset+ansiGreen+ansiBold+"%-12s"+ansiReset+ansiGray+" (экономия %s при буфере %d%%)"+ansiReset+"\n",
				formatCPUValue(optimalLim), formatCPUValue(savings), bufferPercent)
		}
	}

	fmt.Println()
	fmt.Println(thin)

	// Память
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  💾 Память" + ansiReset)
	fmt.Printf(ansiGray+"     Фактически:    "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatMemoryValue(cluster.TotalMemActual))
	fmt.Printf(ansiGray+"     Requests:      "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatMemoryValue(cluster.TotalMemRequest))
	if cluster.TotalMemLimit > 0 {
		memLimRatio := (cluster.TotalMemActual / cluster.TotalMemLimit) * 100.0
		limColor := ansiWhite
		if memLimRatio >= 85.0 {
			limColor = ansiRed
		} else if memLimRatio >= 70.0 {
			limColor = ansiYellow
		}
		fmt.Printf(ansiGray+"     Limits:        "+ansiReset+ansiWhite+"%-12s"+ansiReset+ansiGray+"  факт: "+ansiReset+limColor+"%.0f%% от limits"+ansiReset+"\n",
			formatMemoryValue(cluster.TotalMemLimit), memLimRatio)
		// Потенциальная оптимизация лимитов с учётом буфера
		optimalLim := cluster.TotalMemActual * bufMul
		if cluster.TotalMemLimit > optimalLim*1.1 {
			savings := cluster.TotalMemLimit - optimalLim
			fmt.Printf(ansiGreen+"     💡 Limits можно снизить до: "+ansiReset+ansiGreen+ansiBold+"%-12s"+ansiReset+ansiGray+" (экономия %s при буфере %d%%)"+ansiReset+"\n",
				formatMemoryValue(optimalLim), formatMemoryValue(savings), bufferPercent)
		} else if memLimRatio >= 85.0 {
			fmt.Printf(ansiRed+"     🚨 Высокое потребление! Рассмотрите увеличение limits или масштабирование"+ansiReset+"\n")
		}
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println()
	fmt.Printf(ansiGray+"  📄 Отчёт:    "+ansiReset+ansiGreen+"%s"+ansiReset+"\n", filename)
	fmt.Printf(ansiGray+"  ⏱️  Время:     "+ansiReset+ansiWhite+"%.2f сек"+ansiReset+"\n\n", elapsed.Seconds())
	fmt.Println(ansiGreen + ansiBold + "  ✅ Анализ завершён!" + ansiReset)
	fmt.Println()
}

