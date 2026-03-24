package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Глобальные параметры режима работы
var (
	prometheusURL      string // URL Prometheus/Thanos
	collectDuration    string // Период анализа ("10m", "1h", "7d", "1w" и т.д.)
	thanosCluster      string // Имя кластера в Thanos (лейбл cluster=)
	thanosClusterLabel string // Название лейбла кластера в Thanos (по умолчанию "cluster")
	prometheusMode     bool   // true когда работаем через Prometheus — не опрашиваем Metrics Server
	insecureSkipTLS    bool   // true — не проверять TLS-сертификат (аналог -k в curl/kubectl)
)

// parseDuration разбирает строку периода с поддержкой суффиксов d (дни) и w (недели),
// которые не поддерживает стандартный time.ParseDuration.
// Примеры: "30m", "2h", "7d", "1w", "2w3d".
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, fmt.Errorf("пустая строка")
	}

	// Быстрый путь: стандартный парсер (h, m, s, ms и т.д.)
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Расширенный парсер: поддержка d и w
	var total time.Duration
	rest := s
	for rest != "" {
		i := 0
		for i < len(rest) && (rest[i] >= '0' && rest[i] <= '9') {
			i++
		}
		if i == 0 {
			return 0, fmt.Errorf("неверный формат периода: %q", s)
		}
		n, err := strconv.Atoi(rest[:i])
		if err != nil {
			return 0, err
		}
		rest = rest[i:]

		j := 0
		for j < len(rest) && (rest[j] < '0' || rest[j] > '9') {
			j++
		}
		if j == 0 {
			return 0, fmt.Errorf("ожидался суффикс после числа %d в %q", n, s)
		}
		suffix := rest[:j]
		rest = rest[j:]

		switch suffix {
		case "w":
			total += time.Duration(n) * 7 * 24 * time.Hour
		case "d":
			total += time.Duration(n) * 24 * time.Hour
		case "h":
			total += time.Duration(n) * time.Hour
		case "m":
			total += time.Duration(n) * time.Minute
		case "s":
			total += time.Duration(n) * time.Second
		default:
			return 0, fmt.Errorf("неизвестный суффикс %q в %q (поддерживаются: w, d, h, m, s)", suffix, s)
		}
	}
	return total, nil
}

func main() {
	parseFlags()

	startTime := time.Now()
	printBanner()

	if !validateEnvironment() {
		return
	}

	clusterSummary := collectClusterData()
	filename := generateExcelReport(clusterSummary)
	printFinalSummary(clusterSummary, startTime, filename)
}

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
	flag.StringVar(&prometheusURL, "prometheus", "", "URL Prometheus/Thanos (например: http://prometheus:9090)")
	flag.StringVar(&prometheusURL, "p", "", "URL Prometheus/Thanos (сокращенно)")
	flag.StringVar(&collectDuration, "duration", "", "Период анализа: исторические данные из Prometheus или живой сбор (например: 10m, 1h)")
	flag.StringVar(&collectDuration, "d", "", "Период анализа (сокращенно)")
	flag.StringVar(&thanosCluster, "cluster", "", "Имя кластера в Thanos (автоопределение если не указано)")
	flag.StringVar(&thanosClusterLabel, "cluster-label", "", "Лейбл кластера в Thanos (автоопределение если не указано, обычно 'cluster')")
	flag.BoolVar(&insecureSkipTLS, "insecure", false, "Не проверять TLS-сертификат (как -k в curl)")
	flag.BoolVar(&insecureSkipTLS, "k", false, "Не проверять TLS-сертификат (сокращенно)")
	flag.Parse()

	bufferPercent = buffer

	if version {
		printVersion()
		os.Exit(0)
	}

	if help {
		printHelp()
		os.Exit(0)
	}
}

func validateEnvironment() bool {
	printStep("🔌 Подключение к Kubernetes API...")

	if !checkKubernetesConnection() {
		printError("❌ Не удалось подключиться к кластеру")
		printError("💡 Проверьте:")
		printError("   • Переменная KUBECONFIG указывает на правильный файл")
		printError("   • Файл ~/.kube/config существует и корректен")
		printError("   • У вас есть доступ к кластеру (kubectl get nodes)")
		return false
	}

	if ctx := getCurrentKubeContext(); ctx != "" {
		printStep(fmt.Sprintf("✅ Кластер: %s", ctx))
	} else {
		printStep("✅ Подключение к кластеру установлено")
	}
	return true
}
