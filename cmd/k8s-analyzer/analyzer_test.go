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
// Рекомендации для пода
// ============================================================================

func TestGeneratePodRecommendation_Optimal(t *testing.T) {
	pod := &PodResource{
		CPURequest: "500m", CPUActual: "300m", CPULimit: "1000m",
		MemoryRequest: "512Mi", MemoryActual: "300Mi", MemoryLimit: "1Gi",
	}
	got := generatePodRecommendation(pod, 58.6, 60.0)
	if !strings.Contains(got, "Оптимально") {
		t.Errorf("ожидалась оптимальная рекомендация, получено %q", got)
	}
}

func TestGeneratePodRecommendation_MemoryOverloaded(t *testing.T) {
	pod := &PodResource{
		CPURequest: "500m", CPUActual: "100m", CPULimit: "1000m",
		MemoryRequest: "256Mi", MemoryActual: "512Mi", MemoryLimit: "1Gi",
	}
	got := generatePodRecommendation(pod, 200.0, 20.0)
	if !strings.Contains(got, "занижена") {
		t.Errorf("ожидалось предупреждение о заниженной памяти, получено %q", got)
	}
}

func TestGeneratePodRecommendation_CPUOverprovisioned(t *testing.T) {
	pod := &PodResource{
		CPURequest: "2000m", CPUActual: "50m", CPULimit: "4000m",
		MemoryRequest: "512Mi", MemoryActual: "300Mi", MemoryLimit: "1Gi",
	}
	got := generatePodRecommendation(pod, 58.6, 2.5)
	if !strings.Contains(got, "завышен") {
		t.Errorf("ожидалось предупреждение о завышенном CPU, получено %q", got)
	}
}

func TestGeneratePodRecommendation_NoLimits(t *testing.T) {
	pod := &PodResource{
		CPURequest: "500m", CPUActual: "300m", CPULimit: "",
		MemoryRequest: "512Mi", MemoryActual: "300Mi", MemoryLimit: "",
	}
	got := generatePodRecommendation(pod, 58.6, 60.0)
	if !strings.Contains(got, "limit") {
		t.Errorf("ожидалась рекомендация по установке limit, получено %q", got)
	}
}

// ============================================================================
// Рекомендуемый CPU
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
// Рекомендуемая память
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
