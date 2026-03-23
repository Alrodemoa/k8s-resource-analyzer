package main

// Модуль интеграции с Prometheus/Thanos и живого сбора метрик.
//
// Поддерживает два режима:
//   - Исторический: запрос готовых данных из Prometheus/Thanos за период
//   - Живой сбор: опрос Metrics Server каждые N секунд в течение заданного времени

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// Структуры ответа Prometheus HTTP API
// ============================================================================

// prometheusResponse - ответ от Prometheus/Thanos API
type prometheusResponse struct {
	Status string             `json:"status"`
	Data   prometheusRespData `json:"data"`
}

// prometheusRespData - секция data ответа
type prometheusRespData struct {
	ResultType string              `json:"resultType"`
	Result     []prometheusMetric  `json:"result"`
}

// prometheusMetric - одна временна́я серия
type prometheusMetric struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"` // [[timestamp, "value"], ...]
}

// ============================================================================
// HTTP-клиент Prometheus
// ============================================================================

// newHTTPClient - создаёт HTTP-клиент с учётом флага --insecure/-k
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipTLS, //nolint:gosec
		},
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

// checkPrometheusConnection - проверка доступности Prometheus/Thanos с детальным выводом ошибок.
// Возвращает true если сервер отвечает, иначе печатает конкретную причину.
func checkPrometheusConnection(promURL string) bool {
	client := newHTTPClient(5 * time.Second)
	base := strings.TrimRight(promURL, "/")

	// Пробуем /-/healthy (Prometheus) затем /api/v1/labels (Thanos)
	endpoints := []string{"/-/healthy", "/api/v1/labels"}
	var lastErr error
	var lastStatus int

	for _, path := range endpoints {
		fullURL := base + path
		printStep(fmt.Sprintf("   → проверяю %s", fullURL))
		resp, err := client.Get(fullURL)
		if err != nil {
			lastErr = err
			// Различаем TLS-ошибку от сетевой
			if strings.Contains(err.Error(), "certificate") ||
				strings.Contains(err.Error(), "tls") ||
				strings.Contains(err.Error(), "x509") {
				printError(fmt.Sprintf("   ❌ Ошибка TLS-сертификата: %v", err))
				printError("   💡 Попробуйте запустить с флагом -k для пропуска проверки сертификата")
			} else if strings.Contains(err.Error(), "connection refused") {
				printError(fmt.Sprintf("   ❌ Соединение отклонено — сервер не запущен или неверный порт: %s", promURL))
			} else if strings.Contains(err.Error(), "no such host") ||
				strings.Contains(err.Error(), "dial") {
				printError(fmt.Sprintf("   ❌ Хост недоступен — проверьте DNS и сетевую доступность: %s", promURL))
			} else {
				printError(fmt.Sprintf("   ❌ Ошибка сети: %v", err))
			}
			continue
		}
		resp.Body.Close()
		if resp.StatusCode < 400 {
			return true
		}
		lastStatus = resp.StatusCode
	}

	if lastErr == nil && lastStatus > 0 {
		printError(fmt.Sprintf("   ❌ Сервер вернул HTTP %d — проверьте URL и права доступа", lastStatus))
	}
	return false
}

// labelValuesResponse - ответ /api/v1/label/{name}/values
type labelValuesResponse struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

// getThanosLabelValues - получение всех значений лейбла из Thanos
func getThanosLabelValues(promURL, labelName string) ([]string, error) {
	apiURL := fmt.Sprintf("%s/api/v1/label/%s/values", strings.TrimRight(promURL, "/"), labelName)
	client := newHTTPClient(10 * time.Second)

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result labelValuesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("статус ответа: %s", result.Status)
	}
	return result.Data, nil
}

// getCurrentKubeContext - возвращает имя текущего kubectl-контекста из kubeconfig.
// Используется как подсказка при поиске совпадения среди кластеров Thanos.
func getCurrentKubeContext() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	kubeconfigPath := home + "/.kube/config"
	if kc := os.Getenv("KUBECONFIG"); kc != "" {
		kubeconfigPath = kc
	}
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return ""
	}
	// Ищем строку "current-context: <имя>" без внешних зависимостей
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "current-context:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "current-context:"))
		}
	}
	return ""
}

// detectThanosCluster - автоопределение имени кластера в Thanos.
//
// Перебирает стандартные лейблы (cluster, k8s_cluster, cluster_name),
// сравнивает значения с именем текущего kubectl-контекста.
// Возвращает (имяЛейбла, имяКластера, ошибка).
//
// Логика:
//  1. Если кластер задан явно (--cluster) — используем его.
//  2. Пробуем найти лейбл, значения которого совпадают с контекстом.
//  3. Если кластер один во всём Thanos — используем его без вопросов.
//  4. Если несколько — печатаем список и просим указать --cluster.
func detectThanosCluster(promURL, currentContext string) (labelName, clusterName string, err error) {
	// Стандартные лейблы для идентификации кластера в Thanos
	candidateLabels := []string{"cluster", "k8s_cluster", "cluster_name", "prometheus_cluster"}

	for _, lbl := range candidateLabels {
		values, e := getThanosLabelValues(promURL, lbl)
		if e != nil || len(values) == 0 {
			continue
		}

		// Случай 1: одно значение — кластер единственный
		if len(values) == 1 {
			return lbl, values[0], nil
		}

		// Случай 2: ищем совпадение с именем текущего контекста
		for _, v := range values {
			if v == currentContext || strings.Contains(v, currentContext) || strings.Contains(currentContext, v) {
				return lbl, v, nil
			}
		}

		// Случай 3: несколько кластеров, совпадения нет — показываем список
		printError(fmt.Sprintf("⚠️  Thanos содержит несколько кластеров (лейбл %q):", lbl))
		for _, v := range values {
			printError(fmt.Sprintf("     • %s", v))
		}
		printError("💡 Укажите нужный кластер: --cluster <имя>")
		return lbl, "", fmt.Errorf("не удалось автоматически определить кластер")
	}

	// Лейблы кластера не найдены — Prometheus без multi-cluster
	return "", "", nil
}

// queryPrometheusRange - запрос диапазона данных (query_range)
func queryPrometheusRange(promURL, query string, start, end time.Time, step time.Duration) ([]prometheusMetric, error) {
	apiURL := strings.TrimRight(promURL, "/") + "/api/v1/query_range"

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(end.Unix(), 10))
	params.Set("step", fmt.Sprintf("%.0f", step.Seconds()))

	client := newHTTPClient(60 * time.Second)
	resp, err := client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		msg := fmt.Sprintf("не удалось подключиться к Prometheus (%s): %v", apiURL, err)
		if strings.Contains(err.Error(), "certificate") || strings.Contains(err.Error(), "x509") {
			msg += " — используйте -k для пропуска TLS-проверки"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	var result prometheusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("ошибка разбора ответа Prometheus: %v", err)
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("Prometheus вернул статус: %s", result.Status)
	}

	return result.Data.Result, nil
}

// ============================================================================
// Исторический режим: запрос из Prometheus/Thanos
// ============================================================================

// getPrometheusHistoricalMetrics - получение исторических метрик из Prometheus/Thanos
// promURL     — адрес Prometheus или Thanos Query
// duration    — глубина выборки (например, 10 минут означает "последние 10 минут")
// namespaces  — список неймспейсов для фильтрации (пусто = все)
// clusterLbl  — имя лейбла кластера (например "cluster"), пусто = не фильтровать
// clusterVal  — значение лейбла кластера (например "prod-eu"), пусто = не фильтровать
func getPrometheusHistoricalMetrics(promURL string, duration time.Duration, namespaces []string, clusterLbl, clusterVal string) map[string]*PodHistory {
	end := time.Now()
	start := end.Add(-duration)

	// Шаг: 30 секунд для коротких периодов, 1 минута для длинных
	step := 30 * time.Second
	if duration > time.Hour {
		step = time.Minute
	}

	// Фильтр по неймспейсам
	nsFilter := ""
	if len(namespaces) > 0 {
		nsFilter = fmt.Sprintf(`,namespace=~"%s"`, strings.Join(namespaces, "|"))
	}

	// Фильтр по кластеру (для Thanos multi-cluster)
	clusterFilter := ""
	if clusterLbl != "" && clusterVal != "" {
		clusterFilter = fmt.Sprintf(`,%s="%s"`, clusterLbl, clusterVal)
	}

	// PromQL: CPU в миллиядрах (rate от счётчика * 1000)
	cpuQuery := fmt.Sprintf(
		`sum by (namespace, pod) (rate(container_cpu_usage_seconds_total{container!="POD",container!=""%s%s}[2m])) * 1000`,
		nsFilter, clusterFilter,
	)

	// PromQL: Память в MiB (working set — реальное потребление без кешей)
	memQuery := fmt.Sprintf(
		`sum by (namespace, pod) (container_memory_working_set_bytes{container!="POD",container!=""%s%s}) / 1048576`,
		nsFilter, clusterFilter,
	)

	histories := make(map[string]*PodHistory)

	// Запрос CPU
	printStep("  📡 Запрос CPU метрик из Prometheus...")
	cpuResults, err := queryPrometheusRange(promURL, cpuQuery, start, end, step)
	if err != nil {
		printError(fmt.Sprintf("⚠️  Prometheus CPU: %v", err))
	} else {
		for _, r := range cpuResults {
			ns := r.Metric["namespace"]
			pod := r.Metric["pod"]
			if ns == "" || pod == "" {
				continue
			}
			hist := getOrCreatePodHistory(histories, ns, pod)
			for _, v := range r.Values {
				val := parsePrometheusValue(v)
				hist.CPUSamples = append(hist.CPUSamples, MetricSample{Value: val})
			}
		}
		printStep(fmt.Sprintf("  ✅ CPU: получено данных по %d подам", len(cpuResults)))
	}

	// Запрос памяти
	printStep("  📡 Запрос Memory метрик из Prometheus...")
	memResults, err := queryPrometheusRange(promURL, memQuery, start, end, step)
	if err != nil {
		printError(fmt.Sprintf("⚠️  Prometheus Memory: %v", err))
	} else {
		for _, r := range memResults {
			ns := r.Metric["namespace"]
			pod := r.Metric["pod"]
			if ns == "" || pod == "" {
				continue
			}
			hist := getOrCreatePodHistory(histories, ns, pod)
			for _, v := range r.Values {
				val := parsePrometheusValue(v)
				hist.MemSamples = append(hist.MemSamples, MetricSample{Value: val})
			}
		}
		printStep(fmt.Sprintf("  ✅ Memory: получено данных по %d подам", len(memResults)))
	}

	// Считаем статистику по каждому поду
	for _, hist := range histories {
		calculatePodHistoryStats(hist)
	}

	return histories
}

// ============================================================================
// Живой режим: сбор метрик через Metrics Server
// ============================================================================

// collectLiveMetrics - живой сбор метрик за указанный период.
// Опрашивает Metrics Server каждые 30 секунд, накапливает семплы,
// после окончания периода считает статистику.
func collectLiveMetrics(duration time.Duration, namespaces []string) map[string]*PodHistory {
	interval := 30 * time.Second
	histories := make(map[string]*PodHistory)

	deadline := time.Now().Add(duration)
	sample := 0

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	printStep(fmt.Sprintf("⏱️  Живой сбор: %v, интервал опроса: %v", duration, interval))
	printStep("   Нажмите Ctrl+C для досрочной остановки и генерации отчёта")
	fmt.Println()

	// Первый семпл немедленно
	collectOneSample(histories, namespaces, &sample, deadline)

	for time.Now().Before(deadline) {
		select {
		case <-ticker.C:
			if time.Now().Before(deadline) {
				collectOneSample(histories, namespaces, &sample, deadline)
			}
		}
	}

	fmt.Println()
	printStep(fmt.Sprintf("✅ Сбор завершён: %d семплов по %d подам", sample, len(histories)))

	// Считаем статистику
	for _, hist := range histories {
		calculatePodHistoryStats(hist)
	}

	return histories
}

// collectOneSample - один опрос метрик по всем неймспейсам
func collectOneSample(histories map[string]*PodHistory, namespaces []string, sample *int, deadline time.Time) {
	*sample++
	remaining := time.Until(deadline).Round(time.Second)
	fmt.Printf("\r  📊 Семпл #%-3d | Осталось: %-8v | Подов: %d    ",
		*sample, remaining, len(histories))

	for _, ns := range namespaces {
		usage := getPodActualUsage(ns)
		pods := getPodResources(ns)

		for _, pod := range pods {
			cpuVal := 0.0
			memVal := 0.0

			if u, ok := usage[pod.Name]; ok {
				cpuVal = parseCPUValue(u["cpu"])
				memVal = parseMemoryValue(u["memory"])
			}

			hist := getOrCreatePodHistory(histories, ns, pod.Name)
			hist.CPUSamples = append(hist.CPUSamples, MetricSample{Value: cpuVal})
			hist.MemSamples = append(hist.MemSamples, MetricSample{Value: memVal})
		}
	}
}

// ============================================================================
// Вспомогательные функции
// ============================================================================

// getOrCreatePodHistory - получение или создание записи истории пода
func getOrCreatePodHistory(histories map[string]*PodHistory, ns, pod string) *PodHistory {
	key := ns + "/" + pod
	if _, ok := histories[key]; !ok {
		histories[key] = &PodHistory{
			Namespace: ns,
			Name:      pod,
		}
	}
	return histories[key]
}

// parsePrometheusValue - извлечение значения из семпла Prometheus [[ts, "val"]]
func parsePrometheusValue(v []interface{}) float64 {
	if len(v) < 2 {
		return 0
	}
	str, ok := v[1].(string)
	if !ok {
		return 0
	}
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return val
}

// calculatePodHistoryStats - расчёт статистики (min/avg/max/p95) по семплам.
// После расчёта сырые семплы освобождаются — они больше не нужны.
func calculatePodHistoryStats(hist *PodHistory) {
	// CPU
	if len(hist.CPUSamples) > 0 {
		vals := extractValues(hist.CPUSamples)
		hist.CPUMin, hist.CPUAvg, hist.CPUMax, hist.CPUP95 = calcStats(vals)
		hist.SampleCount = len(hist.CPUSamples)
		hist.CPUSamples = nil // освобождаем память
	}
	// Память
	if len(hist.MemSamples) > 0 {
		vals := extractValues(hist.MemSamples)
		hist.MemMin, hist.MemAvg, hist.MemMax, hist.MemP95 = calcStats(vals)
		hist.MemSamples = nil // освобождаем память
	}
}

// extractValues - извлечение числовых значений из семплов
func extractValues(samples []MetricSample) []float64 {
	vals := make([]float64, len(samples))
	for i, s := range samples {
		vals[i] = s.Value
	}
	return vals
}

// calcStats - расчёт min, avg, max, p95 по набору значений
func calcStats(values []float64) (min, avg, max, p95 float64) {
	if len(values) == 0 {
		return
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	min = sorted[0]
	max = sorted[len(sorted)-1]

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	avg = sum / float64(len(values))

	// P95: 95-й перцентиль
	idx := int(math.Ceil(0.95*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	p95 = sorted[idx]

	return
}
