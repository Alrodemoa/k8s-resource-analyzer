package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

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

func printBanner() {
	sep := ansiCyan + strings.Repeat("─", ConsoleWidth) + ansiReset

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

func printVersion() {
	fmt.Printf("k8s-analyzer v%s\n", AppVersion)
}

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

func printStep(message string) {
	fmt.Println(message)
}

func printError(message string) {
	fmt.Println(ansiRed + message + ansiReset)
}

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

	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  💻 CPU" + ansiReset)
	fmt.Printf(ansiGray+"     Фактически:    "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPUActual))
	fmt.Printf(ansiGray+"     Requests:      "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPURequest))
	if cluster.TotalCPULimit > 0 {
		cpuLimRatio := (cluster.TotalCPUActual / cluster.TotalCPULimit) * 100.0
		fmt.Printf(ansiGray+"     Limits:        "+ansiReset+ansiWhite+"%-12s"+ansiReset+ansiGray+"  факт: %.0f%% от limits"+ansiReset+"\n",
			formatCPUValue(cluster.TotalCPULimit), cpuLimRatio)
		optimalLim := cluster.TotalCPUActual * bufMul
		if cluster.TotalCPULimit > optimalLim*1.1 {
			savings := cluster.TotalCPULimit - optimalLim
			fmt.Printf(ansiGreen+"     💡 Limits можно снизить до: "+ansiReset+ansiGreen+ansiBold+"%-12s"+ansiReset+ansiGray+" (экономия %s при буфере %d%%)"+ansiReset+"\n",
				formatCPUValue(optimalLim), formatCPUValue(savings), bufferPercent)
		}
	}

	fmt.Println()
	fmt.Println(thin)

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

type prometheusFailureAction int

const (
	actionSkipPrometheus prometheusFailureAction = iota
	actionRetryNewURL
	actionRetryInsecure
	actionExit
)

// askPrometheusRetry - интерактивное меню при невозможности подключиться к Prometheus.
// Показывает детальный лог ошибок по каждому проверенному endpoint.
// Возвращает выбранное действие и (опционально) новый URL.
func askPrometheusRetry(failedURL string, probes []ProbeResult) (action prometheusFailureAction, newURL string, newDuration string) {
	sep := ansiCyan + strings.Repeat("─", ConsoleWidth) + ansiReset
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println(sep)
		fmt.Println(ansiYellow + ansiBold + "  ⚠️  Не удалось подключиться к Prometheus" + ansiReset)
		fmt.Println(ansiGray + "     " + failedURL + ansiReset)
		fmt.Println()
		fmt.Println(ansiWhite + "  Лог проверки:" + ansiReset)
		for _, p := range probes {
			if p.OK {
				fmt.Printf(ansiGreen+"  ✅  %s"+ansiReset+" → HTTP %d\n", p.URL, p.Status)
			} else if p.Status > 0 {
				fmt.Printf(ansiRed+"  ✗   %s"+ansiReset+ansiGray+" → %s"+ansiReset+"\n", p.URL, p.Err)
			} else {
				fmt.Printf(ansiRed+"  ✗   %s"+ansiReset+"\n", p.URL)
				fmt.Printf(ansiGray+"       %s"+ansiReset+"\n", p.Err)
			}
		}
		for _, p := range probes {
			if strings.Contains(p.Err, "TLS") || strings.Contains(p.Err, "сертификат") {
				fmt.Println()
				fmt.Println(ansiYellow + "  💡 Обнаружена TLS-ошибка — попробуйте вариант [3] или [4]" + ansiReset)
				break
			}
		}
		fmt.Println(sep)
		fmt.Println()
		fmt.Println(ansiWhite + "  Что делать?" + ansiReset)
		fmt.Println()
		fmt.Println(ansiGreen + "  [1]" + ansiReset + " Продолжить без Prometheus " + ansiGray + "(использовать Metrics Server)" + ansiReset)
		fmt.Println(ansiGreen + "  [2]" + ansiReset + " Указать другой URL Prometheus/Thanos")
		fmt.Println(ansiGreen + "  [3]" + ansiReset + " Повторить с " + ansiYellow + "-k" + ansiReset + " " + ansiGray + "(пропустить проверку TLS-сертификата)" + ansiReset)
		fmt.Println(ansiGreen + "  [4]" + ansiReset + " Указать новый URL " + ansiYellow + "и" + ansiReset + " включить " + ansiYellow + "-k" + ansiReset)
		fmt.Println(ansiRed + "  [0]" + ansiReset + " Выйти")
		fmt.Println()
		fmt.Print(ansiCyan + "  Выбор: " + ansiReset)

		line, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(line)

		switch choice {
		case "1":
			return actionSkipPrometheus, "", ""

		case "2":
			fmt.Print(ansiCyan + "  Новый URL Prometheus/Thanos: " + ansiReset)
			urlLine, _ := reader.ReadString('\n')
			entered := strings.TrimSpace(urlLine)
			if entered == "" {
				printError("  URL не может быть пустым, попробуйте снова")
				continue
			}
			fmt.Print(ansiCyan + "  Период анализа " + ansiGray + "(например: 1h, 7d; Enter = оставить текущий)" + ansiReset + ": ")
			durLine, _ := reader.ReadString('\n')
			return actionRetryNewURL, entered, strings.TrimSpace(durLine)

		case "3":
			return actionRetryInsecure, failedURL, ""

		case "4":
			fmt.Print(ansiCyan + "  Новый URL Prometheus/Thanos: " + ansiReset)
			urlLine, _ := reader.ReadString('\n')
			entered := strings.TrimSpace(urlLine)
			if entered == "" {
				printError("  URL не может быть пустым, попробуйте снова")
				continue
			}
			fmt.Print(ansiCyan + "  Период анализа " + ansiGray + "(например: 1h, 7d; Enter = оставить текущий)" + ansiReset + ": ")
			durLine, _ := reader.ReadString('\n')
			insecureSkipTLS = true
			return actionRetryNewURL, entered, strings.TrimSpace(durLine)

		case "0":
			return actionExit, "", ""

		default:
			printError("  Неверный выбор, введите цифру от 0 до 4")
		}
	}
}

