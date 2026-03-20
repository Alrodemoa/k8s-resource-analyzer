package main

import (
	"strings"
	"testing"
)

// ============================================================================
// Эффективность памяти
// ============================================================================

func TestCalculateMemoryEfficiency_Normal(t *testing.T) {
	pod := &PodResource{MemoryRequest: "512Mi", MemoryActual: "256Mi"}
	got := calculateMemoryEfficiency(pod)
	want := 50.0
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestCalculateMemoryEfficiency_Full(t *testing.T) {
	pod := &PodResource{MemoryRequest: "1Gi", MemoryActual: "1Gi"}
	got := calculateMemoryEfficiency(pod)
	want := 100.0
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestCalculateMemoryEfficiency_ZeroRequest(t *testing.T) {
	pod := &PodResource{MemoryRequest: "0", MemoryActual: "256Mi"}
	got := calculateMemoryEfficiency(pod)
	if got != 0 {
		t.Errorf("ожидалось 0 при нулевом запросе, получено %.2f", got)
	}
}

func TestCalculateMemoryEfficiency_NoData(t *testing.T) {
	pod := &PodResource{MemoryRequest: "Н/Д", MemoryActual: "Н/Д"}
	got := calculateMemoryEfficiency(pod)
	if got != 0 {
		t.Errorf("ожидалось 0 для значений Н/Д, получено %.2f", got)
	}
}

func TestCalculateMemoryEfficiency_OverRequest(t *testing.T) {
	pod := &PodResource{MemoryRequest: "256Mi", MemoryActual: "512Mi"}
	got := calculateMemoryEfficiency(pod)
	want := 200.0
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

// ============================================================================
// Эффективность CPU
// ============================================================================

func TestCalculateCPUEfficiency_Normal(t *testing.T) {
	pod := &PodResource{CPURequest: "500m", CPUActual: "250m"}
	got := calculateCPUEfficiency(pod)
	want := 50.0
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestCalculateCPUEfficiency_Full(t *testing.T) {
	pod := &PodResource{CPURequest: "1000m", CPUActual: "1000m"}
	got := calculateCPUEfficiency(pod)
	want := 100.0
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestCalculateCPUEfficiency_ZeroRequest(t *testing.T) {
	pod := &PodResource{CPURequest: "", CPUActual: "100m"}
	got := calculateCPUEfficiency(pod)
	if got != 0 {
		t.Errorf("ожидалось 0 при нулевом запросе, получено %.2f", got)
	}
}

func TestCalculateCPUEfficiency_ZeroActual(t *testing.T) {
	pod := &PodResource{CPURequest: "500m", CPUActual: "Н/Д"}
	got := calculateCPUEfficiency(pod)
	if got != 0 {
		t.Errorf("ожидалось 0 для фактического Н/Д, получено %.2f", got)
	}
}

// ============================================================================
// Статус пода
// ============================================================================

func TestDeterminePodStatus_Critical(t *testing.T) {
	got := determinePodStatus(105.0, 50.0)
	if !strings.Contains(got, "Критично") {
		t.Errorf("ожидался критический статус, получено %q", got)
	}
}

func TestDeterminePodStatus_High(t *testing.T) {
	got := determinePodStatus(85.0, 40.0)
	if !strings.Contains(got, "Высокая") {
		t.Errorf("ожидался высокий статус, получено %q", got)
	}
}

func TestDeterminePodStatus_Normal(t *testing.T) {
	got := determinePodStatus(60.0, 55.0)
	if !strings.Contains(got, "Оптимально") {
		t.Errorf("ожидался оптимальный статус, получено %q", got)
	}
}

func TestDeterminePodStatus_Low(t *testing.T) {
	got := determinePodStatus(35.0, 32.0)
	if !strings.Contains(got, "Недогруз") {
		t.Errorf("ожидался статус недогруза, получено %q", got)
	}
}

func TestDeterminePodStatus_Minimal(t *testing.T) {
	got := determinePodStatus(5.0, 10.0)
	if !strings.Contains(got, "Минимальная") {
		t.Errorf("ожидался статус минимальной загрузки, получено %q", got)
	}
}

func TestDeterminePodStatus_UsesMaxOfTwoValues(t *testing.T) {
	// память=20 (минимум), cpu=85 (высокий) → должен взять максимум = высокий
	got := determinePodStatus(20.0, 85.0)
	if !strings.Contains(got, "Высокая") {
		t.Errorf("ожидался высокий статус (максимум из двух значений), получено %q", got)
	}
}

// ============================================================================
// Рекомендуемый лимит (calculateRecommendedLimit)
// ============================================================================

func TestCalculateRecommendedLimit_CPU(t *testing.T) {
	bufferPercent = 30
	got := calculateRecommendedLimit("100m", false)
	if got != "130m" {
		t.Errorf("получено %q, ожидалось \"130m\"", got)
	}
}

func TestCalculateRecommendedLimit_Memory(t *testing.T) {
	bufferPercent = 50
	got := calculateRecommendedLimit("1Gi", true)
	// 1024 MiB * 1.5 = 1536 MiB = 1.5 GiB
	if got != "1.5Gi" {
		t.Errorf("получено %q, ожидалось \"1.5Gi\"", got)
	}
}

func TestCalculateRecommendedLimit_EmptyActual(t *testing.T) {
	bufferPercent = 30
	got := calculateRecommendedLimit("Н/Д", true)
	if got != "" {
		t.Errorf("ожидалась пустая строка при отсутствии данных, получено %q", got)
	}
}

func TestCalculateRecommendedLimit_ZeroBuffer(t *testing.T) {
	bufferPercent = 0
	got := calculateRecommendedLimit("200m", false)
	// буфер 0% → рекомендуется = фактическое
	if got != "200m" {
		t.Errorf("получено %q, ожидалось \"200m\"", got)
	}
}

// ============================================================================
// Рекомендации для пода — сводная строка
// ============================================================================

func TestGeneratePodRecommendation_SummaryLinesPresent(t *testing.T) {
	bufferPercent = 30
	pod := &PodResource{
		CPURequest:    "500m",
		CPUActual:     "300m",
		CPULimit:      "1000m",
		MemoryRequest: "512Mi",
		MemoryActual:  "300Mi",
		MemoryLimit:   "1Gi",
	}
	got := generatePodRecommendation(pod, 58.6, 60.0)
	// Сводная строка CPU и MEM должна присутствовать всегда
	if !strings.Contains(got, "CPU  | факт:") {
		t.Errorf("ожидалась сводная строка CPU, получено %q", got)
	}
	if !strings.Contains(got, "MEM  | факт:") {
		t.Errorf("ожидалась сводная строка MEM, получено %q", got)
	}
}

func TestGeneratePodRecommendation_LimitCanBeReduced(t *testing.T) {
	bufferPercent = 30
	// CPU actual=100m, limit=2000m → 100*1.3=130m, 2000 > 130*1.1 → можно снизить
	pod := &PodResource{
		CPURequest:    "200m",
		CPUActual:     "100m",
		CPULimit:      "2000m",
		MemoryRequest: "256Mi",
		MemoryActual:  "128Mi",
		MemoryLimit:   "2Gi",
	}
	got := generatePodRecommendation(pod, 50.0, 50.0)
	if !strings.Contains(got, "можно снизить") {
		t.Errorf("ожидалась рекомендация о снижении limit, получено %q", got)
	}
}

func TestGeneratePodRecommendation_LimitNeedsIncrease(t *testing.T) {
	bufferPercent = 30
	// CPU actual=950m, limit=1000m → 95% ≥ 90% → нужно повысить
	pod := &PodResource{
		CPURequest:    "500m",
		CPUActual:     "950m",
		CPULimit:      "1000m",
		MemoryRequest: "256Mi",
		MemoryActual:  "128Mi",
		MemoryLimit:   "1Gi",
	}
	got := generatePodRecommendation(pod, 50.0, 95.0)
	if !strings.Contains(got, "нужно повысить") {
		t.Errorf("ожидалась рекомендация о повышении limit, получено %q", got)
	}
}

func TestGeneratePodRecommendation_NoLimits(t *testing.T) {
	bufferPercent = 30
	pod := &PodResource{
		CPURequest:    "500m",
		CPUActual:     "300m",
		CPULimit:      "",
		MemoryRequest: "512Mi",
		MemoryActual:  "300Mi",
		MemoryLimit:   "",
	}
	got := generatePodRecommendation(pod, 58.6, 60.0)
	if !strings.Contains(got, "limit не задан") {
		t.Errorf("ожидалось предупреждение об отсутствии limit, получено %q", got)
	}
}

func TestGeneratePodRecommendation_NoActualData(t *testing.T) {
	bufferPercent = 30
	pod := &PodResource{
		CPURequest:    "500m",
		CPUActual:     "Н/Д",
		CPULimit:      "1000m",
		MemoryRequest: "512Mi",
		MemoryActual:  "Н/Д",
		MemoryLimit:   "1Gi",
	}
	got := generatePodRecommendation(pod, 0, 0)
	// Без фактических данных сводные строки не генерируются, лимиты не анализируются
	if strings.Contains(got, "факт:") {
		t.Errorf("не ожидалась сводная строка без фактических данных, получено %q", got)
	}
}

// ============================================================================
// Анализ рекомендаций по памяти
// ============================================================================

func TestAnalyzeMemoryRecommendations_OOMKillRisk(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=920Mi, limit=1000Mi → 92% ≥ 90% → OOMKill риск
	analyzeMemoryRecommendations(&rec, 0, 256, 920, 1000, "256Mi")
	if !strings.Contains(rec.String(), "OOMKill") {
		t.Errorf("ожидался OOMKill риск, получено %q", rec.String())
	}
}

func TestAnalyzeMemoryRecommendations_BelowBuffer(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=800Mi, limit=1000Mi → 80%, bufferMul=1.3, targetRatio=76.9% → 80 ≥ 76.9 → предупреждение
	analyzeMemoryRecommendations(&rec, 0, 256, 800, 1000, "256Mi")
	got := rec.String()
	if !strings.Contains(got, "буфера") {
		t.Errorf("ожидалось предупреждение о запасе ниже буфера, получено %q", got)
	}
}

func TestAnalyzeMemoryRecommendations_LimitCanBeReduced(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=100Mi, limit=1000Mi → recLim=130Mi, 1000 > 130*1.5=195 → можно снизить
	analyzeMemoryRecommendations(&rec, 0, 256, 100, 1000, "256Mi")
	if !strings.Contains(rec.String(), "можно снизить") {
		t.Errorf("ожидалась рекомендация о снижении limit, получено %q", rec.String())
	}
}

func TestAnalyzeMemoryRecommendations_NoLimit(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	analyzeMemoryRecommendations(&rec, 0, 256, 300, 0, "256Mi")
	if !strings.Contains(rec.String(), "limit не задан") {
		t.Errorf("ожидалось предупреждение об отсутствии limit, получено %q", rec.String())
	}
}

func TestAnalyzeMemoryRecommendations_RequestTooHigh(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request=1024Mi, actual=100Mi → request > actual*2 → request завышен
	analyzeMemoryRecommendations(&rec, 0, 1024, 100, 2048, "1Gi")
	if !strings.Contains(rec.String(), "request завышен") {
		t.Errorf("ожидалось предупреждение о завышенном request, получено %q", rec.String())
	}
}

func TestAnalyzeMemoryRecommendations_RequestTooLow(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request=100Mi, actual=500Mi → actual > request*2 → request занижен
	analyzeMemoryRecommendations(&rec, 0, 100, 500, 2048, "100Mi")
	if !strings.Contains(rec.String(), "request занижен") {
		t.Errorf("ожидалось предупреждение о заниженном request, получено %q", rec.String())
	}
}

func TestAnalyzeMemoryRecommendations_InvalidConfig(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request > limit → невалидная конфигурация
	analyzeMemoryRecommendations(&rec, 0, 2048, 300, 1024, "2Gi")
	if !strings.Contains(strings.ToLower(rec.String()), "невалидна") {
		t.Errorf("ожидалась ошибка невалидной конфигурации, получено %q", rec.String())
	}
}

// ============================================================================
// Анализ рекомендаций по CPU
// ============================================================================

func TestAnalyzeCPURecommendations_Throttling(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=950m, limit=1000m → 95% ≥ 90% → throttling
	analyzeCPURecommendations(&rec, 0, 500, 950, 1000, "500m")
	if !strings.Contains(rec.String(), "throttling") {
		t.Errorf("ожидалось предупреждение о throttling, получено %q", rec.String())
	}
}

func TestAnalyzeCPURecommendations_BelowBuffer(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=800m, limit=1000m → 80%, targetRatio=76.9% → предупреждение
	analyzeCPURecommendations(&rec, 0, 500, 800, 1000, "500m")
	got := rec.String()
	if !strings.Contains(got, "буфера") {
		t.Errorf("ожидалось предупреждение о запасе ниже буфера, получено %q", got)
	}
}

func TestAnalyzeCPURecommendations_LimitCanBeReduced(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// actual=50m, limit=2000m → recLim=65m, 2000 > 65*1.5=97.5 → можно снизить
	analyzeCPURecommendations(&rec, 0, 500, 50, 2000, "500m")
	if !strings.Contains(rec.String(), "можно снизить") {
		t.Errorf("ожидалась рекомендация о снижении limit, получено %q", rec.String())
	}
}

func TestAnalyzeCPURecommendations_NoLimit(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	analyzeCPURecommendations(&rec, 0, 500, 300, 0, "500m")
	if !strings.Contains(rec.String(), "limit не задан") {
		t.Errorf("ожидалось предупреждение об отсутствии limit, получено %q", rec.String())
	}
}

func TestAnalyzeCPURecommendations_RequestTooHigh(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request=2000m, actual=50m → request > actual*2 → request завышен
	analyzeCPURecommendations(&rec, 0, 2000, 50, 4000, "2000m")
	if !strings.Contains(rec.String(), "request завышен") {
		t.Errorf("ожидалось предупреждение о завышенном request, получено %q", rec.String())
	}
}

func TestAnalyzeCPURecommendations_RequestTooLow(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request=50m, actual=500m → actual > request*2 → request занижен
	analyzeCPURecommendations(&rec, 0, 50, 500, 2000, "50m")
	if !strings.Contains(rec.String(), "request занижен") {
		t.Errorf("ожидалось предупреждение о заниженном request, получено %q", rec.String())
	}
}

func TestAnalyzeCPURecommendations_InvalidConfig(t *testing.T) {
	bufferPercent = 30
	var rec strings.Builder
	// request=2000m, limit=1000m → невалидная конфигурация
	analyzeCPURecommendations(&rec, 0, 2000, 300, 1000, "2000m")
	if !strings.Contains(strings.ToLower(rec.String()), "невалидна") {
		t.Errorf("ожидалась ошибка невалидной конфигурации, получено %q", rec.String())
	}
}

// ============================================================================
// Буфер влияет на рекомендации
// ============================================================================

func TestBufferPercent_AffectsRecommendedLimit(t *testing.T) {
	// при buffer=0 рекомендуемый лимит = фактическое
	bufferPercent = 0
	got0 := calculateRecommendedLimit("100m", false)

	// при buffer=100 рекомендуемый лимит = фактическое * 2
	bufferPercent = 100
	got100 := calculateRecommendedLimit("100m", false)

	if got0 != "100m" {
		t.Errorf("buffer=0: получено %q, ожидалось \"100m\"", got0)
	}
	if got100 != "200m" {
		t.Errorf("buffer=100: получено %q, ожидалось \"200m\"", got100)
	}
}

func TestBufferPercent_ThresholdChanges(t *testing.T) {
	// При buffer=30%: targetRatio = 1/1.3 * 100 = 76.9%
	// actual=80Mi из limit=100Mi = 80% → выше 76.9% → предупреждение
	bufferPercent = 30
	var rec30 strings.Builder
	analyzeMemoryRecommendations(&rec30, 0, 50, 80, 100, "50Mi")
	if !strings.Contains(rec30.String(), "буфера") {
		t.Errorf("buffer=30: ожидалось предупреждение при 80%% от лимита, получено %q", rec30.String())
	}

	// При buffer=5%: targetRatio = 1/1.05 * 100 = 95.2%
	// actual=80Mi из limit=100Mi = 80% → ниже 95.2% → нет предупреждения (только можно снизить)
	bufferPercent = 5
	var rec5 strings.Builder
	analyzeMemoryRecommendations(&rec5, 0, 50, 80, 100, "50Mi")
	if strings.Contains(rec5.String(), "Запас") {
		t.Errorf("buffer=5: не ожидалось предупреждение при 80%% от лимита, получено %q", rec5.String())
	}
}

// ============================================================================
// Рекомендуемый CPU request
// ============================================================================

func TestCalculateRecommendedCPU_HighEfficiency(t *testing.T) {
	bufferPercent = 50
	// 85% эффективность → берём фактическое * SafetyMarginNormal (1.2)
	pod := &PodResource{CPURequest: "500m", CPUActual: "425m"}
	got := calculateRecommendedCPU(pod)
	// 425 * 1.2 = 510m
	if got != "510m" {
		t.Errorf("получено %q, ожидалось \"510m\"", got)
	}
}

func TestCalculateRecommendedCPU_LowEfficiency(t *testing.T) {
	bufferPercent = 50
	// 10% эффективность → берём фактическое * SafetyMarginUnderutilized (1.3)
	pod := &PodResource{CPURequest: "1000m", CPUActual: "100m"}
	got := calculateRecommendedCPU(pod)
	// 100 * 1.3 = 130m
	if got != "130m" {
		t.Errorf("получено %q, ожидалось \"130m\"", got)
	}
}

func TestCalculateRecommendedCPU_NormalEfficiency(t *testing.T) {
	bufferPercent = 50
	// 60% эффективность → берём запрос * (1 + буфер/100)
	pod := &PodResource{CPURequest: "500m", CPUActual: "300m"}
	got := calculateRecommendedCPU(pod)
	// 500 * 1.5 = 750m
	if got != "750m" {
		t.Errorf("получено %q, ожидалось \"750m\"", got)
	}
}

func TestCalculateRecommendedCPU_NoData(t *testing.T) {
	pod := &PodResource{CPURequest: "500m", CPUActual: "Н/Д"}
	got := calculateRecommendedCPU(pod)
	if got != "500m" {
		t.Errorf("ожидался исходный запрос при отсутствии фактических данных, получено %q", got)
	}
}

// ============================================================================
// Рекомендуемая память request
// ============================================================================

func TestCalculateRecommendedMemory_HighEfficiency(t *testing.T) {
	bufferPercent = 50
	// 90% эффективность → фактическое * SafetyMarginNormal (1.2)
	pod := &PodResource{MemoryRequest: "512Mi", MemoryActual: "461Mi"}
	got := calculateRecommendedMemory(pod)
	// 461 * 1.2 = 553.2Mi → "553Mi"
	if got != "553Mi" {
		t.Errorf("получено %q, ожидалось \"553Mi\"", got)
	}
}

func TestCalculateRecommendedMemory_LowEfficiency(t *testing.T) {
	bufferPercent = 50
	// 10% эффективность → фактическое * SafetyMarginUnderutilized (1.3)
	pod := &PodResource{MemoryRequest: "1Gi", MemoryActual: "102Mi"}
	got := calculateRecommendedMemory(pod)
	// 102 * 1.3 = 132.6Mi → "133Mi"
	if got != "133Mi" {
		t.Errorf("получено %q, ожидалось \"133Mi\"", got)
	}
}

func TestCalculateRecommendedMemory_NoData(t *testing.T) {
	pod := &PodResource{MemoryRequest: "256Mi", MemoryActual: "Н/Д"}
	got := calculateRecommendedMemory(pod)
	if got != "256Mi" {
		t.Errorf("ожидался исходный запрос при отсутствии фактических данных, получено %q", got)
	}
}
