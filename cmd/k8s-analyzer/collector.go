package main

// Модуль сбора данных из Kubernetes кластера

import (
	"fmt"
)

// podHistories - глобальное хранилище исторических данных (заполняется при -d / --duration)
var podHistories map[string]*PodHistory

// collectClusterData - главная функция сбора данных кластера
func collectClusterData() *ClusterSummary {
	printStep("📊 Анализ Kubernetes кластера...")

	selectedNamespaces := getAllNamespaces()
	if len(selectedNamespaces) == 0 {
		printError("❌ Не удалось получить список неймспейсов")
		return &ClusterSummary{}
	}

	// Собираем информацию о нодах
	printStep("🖥️  Сбор информации о нодах...")
	nodes := getNodesInfo()
	printStep(fmt.Sprintf("✅ Найдено нод: %d", len(nodes)))

	// Собираем информацию о PVC и PV
	printStep("💽 Сбор информации о хранилищах...")
	pvcs := getPVCsInfo()
	pvs := getPVsInfo()
	printStep(fmt.Sprintf("✅ Найдено PVC: %d, PV: %d", len(pvcs), len(pvs)))

	// Сбор исторических данных (если задан период --duration)
	if collectDuration != "" {
		duration, err := parseDuration(collectDuration)
		if err != nil {
			printError(fmt.Sprintf("❌ Неверный формат периода: %s (используйте: 30m, 2h, 7d, 1w)", collectDuration))
		} else {
			if prometheusURL != "" {
				// Режим: исторические данные из Prometheus/Thanos
				printStep(fmt.Sprintf("📡 Режим: исторический анализ из Prometheus (%s) за %s", prometheusURL, collectDuration))
				if !checkPrometheusConnection(prometheusURL) {
					printError(fmt.Sprintf("❌ Не удалось подключиться к Prometheus: %s", prometheusURL))
					printError("   Продолжаем с текущими метриками Metrics Server")
				} else {
					printStep("✅ Подключение к Prometheus установлено")

					// Автоопределение кластера в Thanos (если не задан явно)
					resolvedLabel := thanosClusterLabel
					resolvedCluster := thanosCluster
					if resolvedCluster == "" {
						currentCtx := getCurrentKubeContext()
						lbl, cls, e := detectThanosCluster(prometheusURL, currentCtx)
						if e != nil {
							// Ошибка уже напечатана внутри detectThanosCluster
							printError("   Продолжаем без фильтрации по кластеру")
						} else if cls != "" {
							resolvedLabel = lbl
							resolvedCluster = cls
							printStep(fmt.Sprintf("✅ Кластер определён автоматически: %s=%s", lbl, cls))
						}
						// cls == "" означает одиночный Prometheus без multi-cluster лейблов
					} else {
						printStep(fmt.Sprintf("✅ Кластер задан явно: %s=%s", resolvedLabel, resolvedCluster))
					}

					podHistories = getPrometheusHistoricalMetrics(prometheusURL, duration, selectedNamespaces, resolvedLabel, resolvedCluster)
					printStep(fmt.Sprintf("✅ Получена история по %d подам", len(podHistories)))
					prometheusMode = true
				}
			} else {
				// Режим: живой сбор через Metrics Server
				printStep(fmt.Sprintf("⏱️  Режим: живой сбор метрик в течение %s", collectDuration))
				podHistories = collectLiveMetrics(duration, selectedNamespaces)
			}
		}
	}

	// Собираем данные по подам
	printStep("📦 Сбор информации о подах...")
	allPods := collectPodsData(selectedNamespaces)

	// Строим сводку кластера
	cluster := buildClusterSummary(nodes, pvcs, pvs, selectedNamespaces)
	
	// Обрабатываем метрики подов
	processPodsMetrics(cluster, allPods)

	// Рассчитываем утилизацию нод
	calculateNodeUtilization(cluster)

	// Собираем данные о Gatekeeper
	printStep("🔒 Анализ политик Gatekeeper...")
	cluster.Gatekeeper = getGatekeeperStatus()
	if cluster.Gatekeeper.Installed {
		printStep(fmt.Sprintf("✅ Gatekeeper обнаружен: %d шаблонов, %d ограничений",
			len(cluster.Gatekeeper.ConstraintTemplates),
			len(cluster.Gatekeeper.Constraints)))
	} else {
		printStep("ℹ️  Gatekeeper не установлен")
	}

	// Собираем данные о правах доступа (RBAC)
	printStep("👥 Сбор информации о правах доступа (RBAC)...")
	cluster.RBACEntries = getRBACEntries()
	printStep(fmt.Sprintf("✅ Найдено привязок ролей: %d", len(cluster.RBACEntries)))

	printStep("✅ Сбор данных завершен")
	
	return cluster
}

// collectPodsData - сбор данных по подам в выбранных неймспейсах
func collectPodsData(namespaces []string) map[string]map[string]*PodResource {
	allPods := make(map[string]map[string]*PodResource)

	for i, ns := range namespaces {
		printStep(fmt.Sprintf("  [%d/%d] Анализ неймспейса: %s", i+1, len(namespaces), ns))
		
		// Получаем ресурсы подов
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
			// Подставляем данные из Prometheus или живого сбора
			histKey := ns + "/" + pod.Name
			if podHistories != nil {
				if hist, ok := podHistories[histKey]; ok && hist.SampleCount > 0 {
					pod.CPUActual = formatCPUValue(hist.CPUAvg)
					pod.MemoryActual = formatMemoryValue(hist.MemAvg)
				}
			}

			// Рассчитываем эффективность
			memEff := calculateMemoryEfficiency(pod)
			cpuEff := calculateCPUEfficiency(pod)
			
			// Генерируем рекомендации
			pod.Recommendation = generatePodRecommendation(pod, memEff, cpuEff)
			pod.Status = determinePodStatus(memEff, cpuEff)
			
			// Рассчитываем рекомендуемые значения (факт + буфер)
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

// buildClusterSummary - построение сводки кластера
func buildClusterSummary(nodes []*NodeInfo, pvcs []*PVCInfo, pvs []*PVInfo, namespaces []string) *ClusterSummary {
	cluster := &ClusterSummary{
		ByNamespace: make(map[string]*NamespaceSummary),
		ByNode:      make(map[string]*NodeInfo),
		ByPVC:       make(map[string]*PVCInfo),
		ByPV:        make(map[string]*PVInfo),
	}

	// Добавляем информацию о нодах
	cluster.TotalNodes = len(nodes)
	for _, node := range nodes {
		cluster.TotalNodeCPUCapacity += node.CPUCapacity
		cluster.TotalNodeMemoryCapacity += node.MemoryCapacity
		cluster.ByNode[node.Name] = node
	}

	// Добавляем информацию о PVC
	cluster.TotalPVCs = len(pvcs)
	for _, pvc := range pvcs {
		cluster.TotalPVCCapacity += parseMemoryValue(pvc.Capacity)
		cluster.TotalPVCUsed += parseMemoryValue(pvc.Used)
		key := fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
		cluster.ByPVC[key] = pvc
	}

	// Добавляем информацию о PV
	cluster.TotalPVs = len(pvs)
	for _, pv := range pvs {
		cluster.TotalPVCapacity += parseMemoryValue(pv.Capacity)
		cluster.TotalPVUsed += parseMemoryValue(pv.Used)
		cluster.ByPV[pv.Name] = pv
	}

	// Инициализируем сводки по неймспейсам
	for _, ns := range namespaces {
		cluster.ByNamespace[ns] = &NamespaceSummary{}
	}

	return cluster
}

// processPodsMetrics - обработка метрик подов
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

			// Парсим значения
			cpuReq := parseCPUValue(pod.CPURequest)
			cpuAct := parseCPUValue(pod.CPUActual)
			cpuLim := parseCPUValue(pod.CPULimit)
			cpuRec := parseCPUValue(pod.RecommendedCPU)
			memReq := parseMemoryValue(pod.MemoryRequest)
			memAct := parseMemoryValue(pod.MemoryActual)
			memLim := parseMemoryValue(pod.MemoryLimit)
			memRec := parseMemoryValue(pod.RecommendedMem)

			// Обновляем общие суммы
			cluster.TotalCPURequest += cpuReq
			cluster.TotalCPUActual += cpuAct
			cluster.TotalCPULimit += cpuLim
			cluster.TotalCPURecommended += cpuRec
			cluster.TotalMemRequest += memReq
			cluster.TotalMemActual += memAct
			cluster.TotalMemLimit += memLim
			cluster.TotalMemRecommended += memRec

			// Обновляем суммы по неймспейсу
			nsSummary.CPURequestTotal += cpuReq
			nsSummary.CPUActualTotal += cpuAct
			nsSummary.CPURecommendedTotal += cpuRec
			nsSummary.MemRequestTotal += memReq
			nsSummary.MemActualTotal += memAct
			nsSummary.MemRecommendedTotal += memRec

			// Обновляем максимумы
			updateMaxPodMetrics(cluster, pod, ns, cpuAct, cpuReq, memAct, memReq)
		}
	}

	// Рассчитываем оптимизированные значения (от фактического использования)
	cluster.TotalCPUOptimized = cluster.TotalCPUActual * (1.0 + float64(bufferPercent)/100.0)
	cluster.TotalMemOptimized = cluster.TotalMemActual * (1.0 + float64(bufferPercent)/100.0)
}

// updateMaxPodMetrics - обновление максимальных метрик подов
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

// calculateNodeUtilization - расчёт утилизации нод
func calculateNodeUtilization(cluster *ClusterSummary) {
	for _, node := range cluster.ByNode {
		// Рассчитываем утилизацию по фактическому использованию
		if node.CPUCapacity > 0 {
			node.CPUUtilization = (node.CPUActual / node.CPUCapacity) * 100.0
		}
		if node.MemoryCapacity > 0 {
			node.MemoryUtilization = (node.MemoryActual / node.MemoryCapacity) * 100.0
		}

		// Рассчитываем утилизацию по requests
		if node.CPUCapacity > 0 {
			node.CPURequestUtil = (node.CPURequests / node.CPUCapacity) * 100.0
		}
		if node.MemoryCapacity > 0 {
			node.MemoryRequestUtil = (node.MemoryRequests / node.MemoryCapacity) * 100.0
		}

		// Генерируем рекомендации
		node.Recommendation = generateNodeRecommendation(node)
	}
}

// generateNodeRecommendation - генерация рекомендаций для ноды
func generateNodeRecommendation(node *NodeInfo) string {
	if node.CPUUtilization > 80 || node.MemoryUtilization > 80 {
		return "⚠️ Высокая загрузка"
	}
	if node.CPUUtilization < 30 && node.MemoryUtilization < 30 {
		return "💡 Недогружена"
	}
	return "✅ Оптимально"
}
