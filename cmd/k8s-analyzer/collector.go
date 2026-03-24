package main

import (
	"fmt"
	"os"
	"time"
)

var podHistories map[string]*PodHistory

// connectAndFetchPrometheus - подключение к Prometheus/Thanos с интерактивным
// восстановлением при сбое: пользователь может сменить URL, включить -k или
// продолжить без Prometheus.
func connectAndFetchPrometheus(targetURL string, duration time.Duration, namespaces []string) map[string]*PodHistory {
	for {
		printStep(fmt.Sprintf("📡 Prometheus: %s (период: %s)", targetURL, collectDuration))

		if ok, probes := checkPrometheusConnection(targetURL); !ok {
			action, newURL, newDur := askPrometheusRetry(targetURL, probes)
			switch action {
			case actionSkipPrometheus:
				printStep("ℹ️  Продолжаем без Prometheus — используем Metrics Server")
				return nil
			case actionExit:
				os.Exit(0)
			case actionRetryInsecure:
				insecureSkipTLS = true
				printStep("🔓 TLS-проверка отключена, повторяю подключение...")
				continue
			case actionRetryNewURL:
				targetURL = newURL
				prometheusURL = newURL
				if newDur != "" {
					collectDuration = newDur
					if d, err := parseDuration(newDur); err == nil {
						duration = d
					}
				}
				continue
			}
		}

		printStep("✅ Подключение к Prometheus установлено")

		// Автоопределение кластера в Thanos
		resolvedLabel := thanosClusterLabel
		resolvedCluster := thanosCluster
		if resolvedCluster == "" {
			ctx := getCurrentKubeContext()
			lbl, cls, e := detectThanosCluster(targetURL, ctx)
			if e != nil {
				printError("   Продолжаем без фильтрации по кластеру")
			} else if cls != "" {
				resolvedLabel = lbl
				resolvedCluster = cls
				printStep(fmt.Sprintf("✅ Кластер Thanos: %s=%s", lbl, cls))
			}
		} else {
			printStep(fmt.Sprintf("✅ Кластер задан явно: %s=%s", resolvedLabel, resolvedCluster))
		}

		histories := getPrometheusHistoricalMetrics(targetURL, duration, namespaces, resolvedLabel, resolvedCluster)
		printStep(fmt.Sprintf("✅ Получена история по %d подам", len(histories)))
		prometheusMode = true
		return histories
	}
}

// collectClusterData - главная функция сбора всех данных кластера.
func collectClusterData() *ClusterSummary {
	printStep("📊 Анализ Kubernetes кластера...")

	selectedNamespaces := getAllNamespaces()
	if len(selectedNamespaces) == 0 {
		printError("❌ Не удалось получить список неймспейсов")
		return &ClusterSummary{}
	}

	printStep("🖥️  Сбор информации о нодах...")
	nodes := getNodesInfo()
	printStep(fmt.Sprintf("✅ Найдено нод: %d", len(nodes)))

	printStep("💽 Сбор информации о хранилищах...")
	pvcs := getPVCsInfo()
	pvs := getPVsInfo()
	printStep(fmt.Sprintf("✅ Найдено PVC: %d, PV: %d", len(pvcs), len(pvs)))

	if collectDuration != "" {
		duration, err := parseDuration(collectDuration)
		if err != nil {
			printError(fmt.Sprintf("❌ Неверный формат периода: %s (используйте: 30m, 2h, 7d, 1w)", collectDuration))
		} else {
			if prometheusURL != "" {
				podHistories = connectAndFetchPrometheus(prometheusURL, duration, selectedNamespaces)
			} else {
				printStep(fmt.Sprintf("⏱️  Режим: живой сбор метрик в течение %s", collectDuration))
				podHistories = collectLiveMetrics(duration, selectedNamespaces)
			}
		}
	}

	printStep("📦 Сбор информации о подах...")
	allPods := collectPodsData(selectedNamespaces)

	cluster := buildClusterSummary(nodes, pvcs, pvs, selectedNamespaces)
	processPodsMetrics(cluster, allPods)
	calculateNodeUtilization(cluster)

	printStep("🔒 Анализ политик Gatekeeper...")
	cluster.Gatekeeper = getGatekeeperStatus()
	if cluster.Gatekeeper.Installed {
		printStep(fmt.Sprintf("✅ Gatekeeper обнаружен: %d шаблонов, %d ограничений",
			len(cluster.Gatekeeper.ConstraintTemplates),
			len(cluster.Gatekeeper.Constraints)))
	} else {
		printStep("ℹ️  Gatekeeper не установлен")
	}

	printStep("👥 Сбор информации о правах доступа (RBAC)...")
	cluster.RBACEntries = getRBACEntries()
	printStep(fmt.Sprintf("✅ Найдено привязок ролей: %d", len(cluster.RBACEntries)))

	printStep("✅ Сбор данных завершен")
	
	return cluster
}

func collectPodsData(namespaces []string) map[string]map[string]*PodResource {
	allPods := make(map[string]map[string]*PodResource)

	for i, ns := range namespaces {
		printStep(fmt.Sprintf("  [%d/%d] Анализ неймспейса: %s", i+1, len(namespaces), ns))

		pods := getPodResources(ns)

		// В Prometheus-режиме Metrics Server не опрашиваем —
		// фактические значения подставятся из podHistories ниже
		if !prometheusMode {
			actualUsage := getPodActualUsage(ns)
			for _, pod := range pods {
				if usage, ok := actualUsage[pod.Name]; ok {
					pod.CPUActual = usage["cpu"]
					pod.MemoryActual = usage["memory"]
				}
			}
		}

		for _, pod := range pods {
			histKey := ns + "/" + pod.Name
			if podHistories != nil {
				if hist, ok := podHistories[histKey]; ok && hist.SampleCount > 0 {
					pod.CPUActual = formatCPUValue(hist.CPUAvg)
					pod.MemoryActual = formatMemoryValue(hist.MemAvg)
				}
			}

			memEff := calculateMemoryEfficiency(pod)
			cpuEff := calculateCPUEfficiency(pod)
			pod.Recommendation = generatePodRecommendation(pod, memEff, cpuEff)
			pod.Status = determinePodStatus(memEff, cpuEff)
			pod.RecommendedCPU = calculateRecommendedCPU(pod)
			pod.RecommendedMem = calculateRecommendedMemory(pod)
			pod.RecommendedCPULimit = calculateRecommendedLimit(pod.CPUActual, false)
			pod.RecommendedMemLimit = calculateRecommendedLimit(pod.MemoryActual, true)
		}
		
		allPods[ns] = make(map[string]*PodResource)
		for _, pod := range pods {
			allPods[ns][pod.Name] = pod
		}
	}

	return allPods
}

func buildClusterSummary(nodes []*NodeInfo, pvcs []*PVCInfo, pvs []*PVInfo, namespaces []string) *ClusterSummary {
	cluster := &ClusterSummary{
		ByNamespace: make(map[string]*NamespaceSummary),
		ByNode:      make(map[string]*NodeInfo),
		ByPVC:       make(map[string]*PVCInfo),
		ByPV:        make(map[string]*PVInfo),
	}

	cluster.TotalNodes = len(nodes)
	for _, node := range nodes {
		cluster.TotalNodeCPUCapacity += node.CPUCapacity
		cluster.TotalNodeMemoryCapacity += node.MemoryCapacity
		cluster.ByNode[node.Name] = node
	}

	cluster.TotalPVCs = len(pvcs)
	for _, pvc := range pvcs {
		cluster.TotalPVCCapacity += parseMemoryValue(pvc.Capacity)
		cluster.TotalPVCUsed += parseMemoryValue(pvc.Used)
		key := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
		cluster.ByPVC[key] = pvc
	}

	cluster.TotalPVs = len(pvs)
	for _, pv := range pvs {
		cluster.TotalPVCapacity += parseMemoryValue(pv.Capacity)
		cluster.TotalPVUsed += parseMemoryValue(pv.Used)
		cluster.ByPV[pv.Name] = pv
	}

	for _, ns := range namespaces {
		cluster.ByNamespace[ns] = &NamespaceSummary{}
	}

	return cluster
}

func processPodsMetrics(cluster *ClusterSummary, allPods map[string]map[string]*PodResource) {
	for ns, pods := range allPods {
		nsSummary := cluster.ByNamespace[ns]
		if nsSummary == nil {
			nsSummary = &NamespaceSummary{}
			cluster.ByNamespace[ns] = nsSummary
		}

		for _, pod := range pods {
			cluster.TotalPods++
			nsSummary.PodCount++

			cpuReq := parseCPUValue(pod.CPURequest)
			cpuAct := parseCPUValue(pod.CPUActual)
			cpuLim := parseCPUValue(pod.CPULimit)
			cpuRec := parseCPUValue(pod.RecommendedCPU)
			memReq := parseMemoryValue(pod.MemoryRequest)
			memAct := parseMemoryValue(pod.MemoryActual)
			memLim := parseMemoryValue(pod.MemoryLimit)
			memRec := parseMemoryValue(pod.RecommendedMem)

			cluster.TotalCPURequest += cpuReq
			cluster.TotalCPUActual += cpuAct
			cluster.TotalCPULimit += cpuLim
			cluster.TotalCPURecommended += cpuRec
			cluster.TotalMemRequest += memReq
			cluster.TotalMemActual += memAct
			cluster.TotalMemLimit += memLim
			cluster.TotalMemRecommended += memRec

			nsSummary.CPURequestTotal += cpuReq
			nsSummary.CPUActualTotal += cpuAct
			nsSummary.CPURecommendedTotal += cpuRec
			nsSummary.MemRequestTotal += memReq
			nsSummary.MemActualTotal += memAct
			nsSummary.MemRecommendedTotal += memRec

			updateMaxPodMetrics(cluster, pod, ns, cpuAct, cpuReq, memAct, memReq)
		}
	}

	cluster.TotalCPUOptimized = cluster.TotalCPUActual * (1.0 + float64(bufferPercent)/100.0)
	cluster.TotalMemOptimized = cluster.TotalMemActual * (1.0 + float64(bufferPercent)/100.0)
}

func updateMaxPodMetrics(cluster *ClusterSummary, pod *PodResource, ns string, cpuAct, cpuReq, memAct, memReq float64) {
	if cpuAct > cluster.MaxPodCPUActual {
		cluster.MaxPodCPUActual = cpuAct
		cluster.MaxPodNameCPU = pod.Name
		cluster.MaxPodNamespaceCPU = ns
	}
	if cpuReq > cluster.MaxPodCPURequest {
		cluster.MaxPodCPURequest = cpuReq
	}
	if memAct > cluster.MaxPodMemoryActual {
		cluster.MaxPodMemoryActual = memAct
		cluster.MaxPodNameMemory = pod.Name
		cluster.MaxPodNamespaceMemory = ns
	}
	if memReq > cluster.MaxPodMemoryRequest {
		cluster.MaxPodMemoryRequest = memReq
	}
}

func calculateNodeUtilization(cluster *ClusterSummary) {
	for _, node := range cluster.ByNode {
		if node.CPUCapacity > 0 {
			node.CPUUtilization = (node.CPUActual / node.CPUCapacity) * 100.0
		}
		if node.MemoryCapacity > 0 {
			node.MemoryUtilization = (node.MemoryActual / node.MemoryCapacity) * 100.0
		}

		if node.CPUCapacity > 0 {
			node.CPURequestUtil = (node.CPURequests / node.CPUCapacity) * 100.0
		}
		if node.MemoryCapacity > 0 {
			node.MemoryRequestUtil = (node.MemoryRequests / node.MemoryCapacity) * 100.0
		}

		node.Recommendation = generateNodeRecommendation(node)
	}
}

func generateNodeRecommendation(node *NodeInfo) string {
	if node.CPUUtilization > 80 || node.MemoryUtilization > 80 {
		return "⚠️ Высокая загрузка"
	}
	if node.CPUUtilization < 30 && node.MemoryUtilization < 30 {
		return "💡 Недогружена"
	}
	return "✅ Оптимально"
}
