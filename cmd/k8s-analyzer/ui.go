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
	fmt.Printf(ansiGray+"  🔧 Запас ресурсов: "+ansiReset+ansiWhite+"%d%%"+ansiReset+"\n\n",
		bufferPercent)
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
	fmt.Println(ansiGreen + "    -b, --buffer <число>" + ansiReset + "   Процент запаса ресурсов " + ansiGray + "(по умолчанию 50)" + ansiReset)
	fmt.Println(ansiGreen + "    -v, --version" + ansiReset + "          Показать версию программы")
	fmt.Println(ansiGreen + "    -h, --help" + ansiReset + "             Показать эту справку")
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  Примеры:" + ansiReset)
	fmt.Println(ansiGray + "    k8s-analyzer" + ansiReset + "              # Запас 50% (по умолчанию)")
	fmt.Println(ansiGray + "    k8s-analyzer -b 30" + ansiReset + "        # Запас 30% (dev)")
	fmt.Println(ansiGray + "    k8s-analyzer -b 100" + ansiReset + "       # Запас 100% (prod)")
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

	// CPU
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  💻 CPU" + ansiReset)
	fmt.Printf(ansiGray+"     Запросы:       "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPURequest))
	fmt.Printf(ansiGray+"     Фактически:    "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPUActual))
	fmt.Printf(ansiGray+"     Рекомендуется: "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatCPUValue(cluster.TotalCPURecommended))

	cpuSavings := cluster.TotalCPURequest - cluster.TotalCPURecommended
	if cpuSavings > 0 {
		fmt.Printf(ansiGreen+"     💰 Экономия:    "+ansiReset+ansiGreen+ansiBold+"%-12s"+ansiReset+ansiGray+" (%.1f%%)"+ansiReset+"\n",
			formatCPUValue(cpuSavings), (cpuSavings/cluster.TotalCPURequest)*100)
	} else {
		fmt.Printf(ansiRed+"     📈 Дефицит:     "+ansiReset+ansiRed+ansiBold+"%-12s"+ansiReset+ansiGray+" (%.1f%%)"+ansiReset+"\n",
			formatCPUValue(-cpuSavings), (-cpuSavings/cluster.TotalCPURequest)*100)
	}

	fmt.Println()
	fmt.Println(thin)

	// Память
	fmt.Println()
	fmt.Println(ansiYellow + ansiBold + "  💾 Память" + ansiReset)
	fmt.Printf(ansiGray+"     Запросы:       "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatMemoryValue(cluster.TotalMemRequest))
	fmt.Printf(ansiGray+"     Фактически:    "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatMemoryValue(cluster.TotalMemActual))
	fmt.Printf(ansiGray+"     Рекомендуется: "+ansiReset+ansiWhite+"%-12s"+ansiReset+"\n", formatMemoryValue(cluster.TotalMemRecommended))

	memSavings := cluster.TotalMemRequest - cluster.TotalMemRecommended
	if memSavings > 0 {
		fmt.Printf(ansiGreen+"     💰 Экономия:    "+ansiReset+ansiGreen+ansiBold+"%-12s"+ansiReset+ansiGray+" (%.1f%%)"+ansiReset+"\n",
			formatMemoryValue(memSavings), (memSavings/cluster.TotalMemRequest)*100)
	} else {
		fmt.Printf(ansiRed+"     📈 Дефицит:     "+ansiReset+ansiRed+ansiBold+"%-12s"+ansiReset+ansiGray+" (%.1f%%)"+ansiReset+"\n",
			formatMemoryValue(-memSavings), (-memSavings/cluster.TotalMemRequest)*100)
	}

	fmt.Println()
	fmt.Println(sep)
	fmt.Println()
	fmt.Printf(ansiGray+"  📄 Отчёт:    "+ansiReset+ansiGreen+"%s"+ansiReset+"\n", filename)
	fmt.Printf(ansiGray+"  ⏱️  Время:     "+ansiReset+ansiWhite+"%.2f сек"+ansiReset+"\n\n", elapsed.Seconds())
	fmt.Println(ansiGreen + ansiBold + "  ✅ Анализ завершён!" + ansiReset)
	fmt.Println()
}

// selectNamespaces - интерактивный выбор неймспейсов
func selectNamespaces(allNamespaces []string) []string {
	sep := ansiCyan + strings.Repeat("═", ConsoleWidth) + ansiReset

	fmt.Println("\n" + sep)
	fmt.Println(centerText(ansiWhite+ansiBold+"  ВЫБОР НЕЙМСПЕЙСОВ  "+ansiReset, ConsoleWidth+len(ansiWhite+ansiBold+ansiReset)))
	fmt.Println(sep)
	fmt.Println()

	for i, ns := range allNamespaces {
		fmt.Printf(ansiGray+"  %2d."+ansiReset+"  "+ansiWhite+"%-40s"+ansiReset+"\n", i+1, ns)
	}

	fmt.Println()
	fmt.Println(ansiYellow + "  Форматы выбора:" + ansiReset)
	fmt.Println(ansiGray + "    1,3,5" + ansiReset + "  — через запятую")
	fmt.Println(ansiGray + "    1-10 " + ansiReset + "  — диапазон")
	fmt.Println(ansiGray + "    *    " + ansiReset + "  — все неймспейсы")
	fmt.Print(ansiCyan + "\n  Ваш выбор: " + ansiReset)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "*" {
		return allNamespaces
	}

	indices := parseSelection(input, len(allNamespaces))
	if len(indices) == 0 {
		return []string{}
	}

	var selected []string
	for _, idx := range indices {
		if idx >= 1 && idx <= len(allNamespaces) {
			selected = append(selected, allNamespaces[idx-1])
		}
	}

	return selected
}
