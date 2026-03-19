package main

import (
	"testing"
)

// ============================================================================
// generateNodeRecommendation
// ============================================================================

func TestGenerateNodeRecommendation_HighCPU(t *testing.T) {
	node := &NodeInfo{CPUUtilization: 85.0, MemoryUtilization: 40.0}
	got := generateNodeRecommendation(node)
	want := "⚠️ Высокая загрузка"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateNodeRecommendation_HighMemory(t *testing.T) {
	node := &NodeInfo{CPUUtilization: 40.0, MemoryUtilization: 90.0}
	got := generateNodeRecommendation(node)
	want := "⚠️ Высокая загрузка"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateNodeRecommendation_BothHigh(t *testing.T) {
	node := &NodeInfo{CPUUtilization: 95.0, MemoryUtilization: 95.0}
	got := generateNodeRecommendation(node)
	want := "⚠️ Высокая загрузка"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateNodeRecommendation_BothLow(t *testing.T) {
	node := &NodeInfo{CPUUtilization: 10.0, MemoryUtilization: 10.0}
	got := generateNodeRecommendation(node)
	want := "💡 Недогружена"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateNodeRecommendation_Optimal(t *testing.T) {
	node := &NodeInfo{CPUUtilization: 55.0, MemoryUtilization: 60.0}
	got := generateNodeRecommendation(node)
	want := "✅ Оптимально"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGenerateNodeRecommendation_LowCPUHighMemory(t *testing.T) {
	// CPU низкий, память высокая → "высокая загрузка" побеждает
	node := &NodeInfo{CPUUtilization: 5.0, MemoryUtilization: 85.0}
	got := generateNodeRecommendation(node)
	want := "⚠️ Высокая загрузка"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ============================================================================
// updateMaxPodMetrics
// ============================================================================

func TestUpdateMaxPodMetrics_UpdatesAll(t *testing.T) {
	cluster := &ClusterSummary{}
	pod := &PodResource{Name: "my-pod"}
	updateMaxPodMetrics(cluster, pod, "default", 500, 600, 1024, 2048)

	if cluster.MaxPodCPUActual != 500 {
		t.Errorf("MaxPodCPUActual = %.0f, want 500", cluster.MaxPodCPUActual)
	}
	if cluster.MaxPodCPURequest != 600 {
		t.Errorf("MaxPodCPURequest = %.0f, want 600", cluster.MaxPodCPURequest)
	}
	if cluster.MaxPodMemoryActual != 1024 {
		t.Errorf("MaxPodMemoryActual = %.0f, want 1024", cluster.MaxPodMemoryActual)
	}
	if cluster.MaxPodMemoryRequest != 2048 {
		t.Errorf("MaxPodMemoryRequest = %.0f, want 2048", cluster.MaxPodMemoryRequest)
	}
	if cluster.MaxPodNameCPU != "my-pod" {
		t.Errorf("MaxPodNameCPU = %q, want \"my-pod\"", cluster.MaxPodNameCPU)
	}
	if cluster.MaxPodNameMemory != "my-pod" {
		t.Errorf("MaxPodNameMemory = %q, want \"my-pod\"", cluster.MaxPodNameMemory)
	}
	if cluster.MaxPodNamespaceCPU != "default" {
		t.Errorf("MaxPodNamespaceCPU = %q, want \"default\"", cluster.MaxPodNamespaceCPU)
	}
}

func TestUpdateMaxPodMetrics_DoesNotDowngrade(t *testing.T) {
	cluster := &ClusterSummary{
		MaxPodCPUActual:    1000,
		MaxPodMemoryActual: 2048,
		MaxPodNameCPU:      "big-pod",
		MaxPodNameMemory:   "big-pod",
	}
	pod := &PodResource{Name: "small-pod"}
	updateMaxPodMetrics(cluster, pod, "default", 100, 100, 100, 100)

	if cluster.MaxPodCPUActual != 1000 {
		t.Errorf("max should not decrease: got %.0f", cluster.MaxPodCPUActual)
	}
	if cluster.MaxPodNameCPU != "big-pod" {
		t.Errorf("name should not change: got %q", cluster.MaxPodNameCPU)
	}
	if cluster.MaxPodMemoryActual != 2048 {
		t.Errorf("max memory should not decrease: got %.0f", cluster.MaxPodMemoryActual)
	}
}

func TestUpdateMaxPodMetrics_UpdatesOnEqual(t *testing.T) {
	cluster := &ClusterSummary{MaxPodCPUActual: 500}
	pod := &PodResource{Name: "equal-pod"}
	// Значение равно текущему максимуму — не должно обновиться (условие >)
	updateMaxPodMetrics(cluster, pod, "ns", 500, 0, 0, 0)
	if cluster.MaxPodNameCPU == "equal-pod" {
		t.Errorf("name should not update on equal value")
	}
}

// ============================================================================
// buildClusterSummary
// ============================================================================

func TestBuildClusterSummary_BasicAggregation(t *testing.T) {
	nodes := []*NodeInfo{
		{Name: "node-1", CPUCapacity: 4000, MemoryCapacity: 8192},
		{Name: "node-2", CPUCapacity: 8000, MemoryCapacity: 16384},
	}
	pvcs := []*PVCInfo{
		{Namespace: "default", Name: "pvc-1", Capacity: "10Gi", Used: "5Gi"},
	}
	pvs := []*PVInfo{
		{Name: "pv-1", Capacity: "10Gi", Used: "5Gi"},
	}
	namespaces := []string{"default", "monitoring"}

	cluster := buildClusterSummary(nodes, pvcs, pvs, namespaces)

	if cluster.TotalNodes != 2 {
		t.Errorf("TotalNodes = %d, want 2", cluster.TotalNodes)
	}
	if cluster.TotalNodeCPUCapacity != 12000 {
		t.Errorf("TotalNodeCPUCapacity = %.0f, want 12000", cluster.TotalNodeCPUCapacity)
	}
	if cluster.TotalNodeMemoryCapacity != 24576 {
		t.Errorf("TotalNodeMemoryCapacity = %.0f, want 24576", cluster.TotalNodeMemoryCapacity)
	}
	if cluster.TotalPVCs != 1 {
		t.Errorf("TotalPVCs = %d, want 1", cluster.TotalPVCs)
	}
	if cluster.TotalPVs != 1 {
		t.Errorf("TotalPVs = %d, want 1", cluster.TotalPVs)
	}
}

func TestBuildClusterSummary_NamespacesInitialized(t *testing.T) {
	namespaces := []string{"default", "kube-system", "monitoring"}
	cluster := buildClusterSummary(nil, nil, nil, namespaces)

	for _, ns := range namespaces {
		if _, ok := cluster.ByNamespace[ns]; !ok {
			t.Errorf("namespace %q not initialized in ByNamespace", ns)
		}
	}
}

func TestBuildClusterSummary_NodesMappedByName(t *testing.T) {
	nodes := []*NodeInfo{
		{Name: "node-1", CPUCapacity: 2000},
		{Name: "node-2", CPUCapacity: 4000},
	}
	cluster := buildClusterSummary(nodes, nil, nil, nil)

	if _, ok := cluster.ByNode["node-1"]; !ok {
		t.Errorf("node-1 not found in ByNode map")
	}
	if _, ok := cluster.ByNode["node-2"]; !ok {
		t.Errorf("node-2 not found in ByNode map")
	}
}

func TestBuildClusterSummary_PVCMappedByNamespaceName(t *testing.T) {
	pvcs := []*PVCInfo{
		{Namespace: "default", Name: "pvc-data", Capacity: "5Gi", Used: "0"},
	}
	cluster := buildClusterSummary(nil, pvcs, nil, nil)

	key := "default/pvc-data"
	if _, ok := cluster.ByPVC[key]; !ok {
		t.Errorf("PVC not found by key %q", key)
	}
}

func TestBuildClusterSummary_EmptyInputs(t *testing.T) {
	cluster := buildClusterSummary(nil, nil, nil, nil)

	if cluster.TotalNodes != 0 || cluster.TotalPVCs != 0 || cluster.TotalPVs != 0 {
		t.Errorf("expected zeros for empty input, got nodes=%d pvcs=%d pvs=%d",
			cluster.TotalNodes, cluster.TotalPVCs, cluster.TotalPVs)
	}
	if cluster.ByNamespace == nil || cluster.ByNode == nil {
		t.Errorf("maps should be initialized, not nil")
	}
}

// ============================================================================
// calculateNodeUtilization
// ============================================================================

func TestCalculateNodeUtilization_Percentages(t *testing.T) {
	node := &NodeInfo{
		Name:           "node-1",
		CPUCapacity:    4000,
		MemoryCapacity: 8192,
		CPUActual:      2000,
		MemoryActual:   4096,
		CPURequests:    3000,
		MemoryRequests: 6144,
	}
	cluster := &ClusterSummary{
		ByNode: map[string]*NodeInfo{"node-1": node},
	}

	calculateNodeUtilization(cluster)

	if node.CPUUtilization != 50.0 {
		t.Errorf("CPUUtilization = %.1f, want 50.0", node.CPUUtilization)
	}
	if node.MemoryUtilization != 50.0 {
		t.Errorf("MemoryUtilization = %.1f, want 50.0", node.MemoryUtilization)
	}
	if node.CPURequestUtil != 75.0 {
		t.Errorf("CPURequestUtil = %.1f, want 75.0", node.CPURequestUtil)
	}
	if node.MemoryRequestUtil != 75.0 {
		t.Errorf("MemoryRequestUtil = %.1f, want 75.0", node.MemoryRequestUtil)
	}
}

func TestCalculateNodeUtilization_ZeroCapacity(t *testing.T) {
	node := &NodeInfo{
		Name:           "node-zero",
		CPUCapacity:    0,
		MemoryCapacity: 0,
		CPUActual:      500,
		MemoryActual:   1024,
	}
	cluster := &ClusterSummary{
		ByNode: map[string]*NodeInfo{"node-zero": node},
	}

	// Не должно паниковать при делении на ноль
	calculateNodeUtilization(cluster)

	if node.CPUUtilization != 0 {
		t.Errorf("expected 0 utilization for zero capacity, got %.2f", node.CPUUtilization)
	}
}

func TestCalculateNodeUtilization_SetsRecommendation(t *testing.T) {
	node := &NodeInfo{
		Name:           "node-1",
		CPUCapacity:    4000,
		MemoryCapacity: 8192,
		CPUActual:      200,
		MemoryActual:   512,
	}
	cluster := &ClusterSummary{
		ByNode: map[string]*NodeInfo{"node-1": node},
	}

	calculateNodeUtilization(cluster)

	if node.Recommendation == "" {
		t.Errorf("expected recommendation to be set, got empty string")
	}
}
