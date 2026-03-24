package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	clientset     *kubernetes.Clientset
	metricsClient *metricsv.Clientset
	dynamicClient dynamic.Interface
)

func initKubernetesClient() error {
	config, err := getKubeConfig()
	if err != nil {
		return fmt.Errorf("не удалось получить конфигурацию: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("не удалось создать клиент: %v", err)
	}

	metricsClient, err = metricsv.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("не удалось создать клиент метрик: %v", err)
	}

	// Dynamic клиент нужен только для Gatekeeper — ошибка не критична
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		dynamicClient = nil
	}

	return nil
}

func getKubeConfig() (*rest.Config, error) {
	// Сначала пробуем in-cluster конфигурацию (если запущено внутри пода)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	var kubeconfig string
	if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
		kubeconfig = envKubeconfig
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить kubeconfig: %v", err)
	}

	return config, nil
}

func checkKubernetesConnection() bool {
	if err := initKubernetesClient(); err != nil {
		return false
	}

	ctx := context.Background()
	_, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	return err == nil
}

func getAllNamespaces() []string {
	ctx := context.Background()
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("Ошибка получения неймспейсов: %v", err))
		return []string{}
	}

	var result []string
	for _, ns := range namespaces.Items {
		result = append(result, ns.Name)
	}
	return result
}

func getNodesInfo() []*NodeInfo {
	ctx := context.Background()
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("Ошибка получения нод: %v", err))
		return []*NodeInfo{}
	}

	var nodes []*NodeInfo
	for _, node := range nodeList.Items {
		nodeInfo := &NodeInfo{
			Name:           node.Name,
			CPUCapacity:    float64(node.Status.Capacity.Cpu().MilliValue()),
			MemoryCapacity: float64(node.Status.Capacity.Memory().Value()) / (1024 * 1024), // в MiB
		}
		nodes = append(nodes, nodeInfo)
	}

	enrichNodesWithMetrics(nodes)

	return nodes
}

func enrichNodesWithMetrics(nodes []*NodeInfo) {
	ctx := context.Background()
	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("⚠️  Предупреждение: не удалось получить метрики нод: %v", err))
		return
	}

	metricsMap := make(map[string]struct {
		cpu    int64
		memory int64
	})
	for _, metric := range nodeMetrics.Items {
		metricsMap[metric.Name] = struct {
			cpu    int64
			memory int64
		}{
			cpu:    metric.Usage.Cpu().MilliValue(),
			memory: metric.Usage.Memory().Value() / (1024 * 1024), // в MiB
		}
	}

	for _, node := range nodes {
		if metrics, ok := metricsMap[node.Name]; ok {
			node.CPUActual = float64(metrics.cpu)
			node.MemoryActual = float64(metrics.memory)
		}
	}
}

func getPodResources(namespace string) []*PodResource {
	ctx := context.Background()
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("Ошибка получения подов в %s: %v", namespace, err))
		return []*PodResource{}
	}

	var pods []*PodResource
	for _, pod := range podList.Items {
		podResource := convertPodToPodResource(&pod, namespace)
		if podResource != nil {
			pods = append(pods, podResource)
		}
	}

	return pods
}

func convertPodToPodResource(pod *corev1.Pod, namespace string) *PodResource {
	podResource := &PodResource{
		Namespace: namespace,
		Name:      pod.Name,
		NodeName:  pod.Spec.NodeName,
		PVCs:      []string{},
	}

	if podResource.NodeName == "" {
		podResource.NodeName = "не назначена"
	}

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			podResource.PVCs = append(podResource.PVCs, volume.PersistentVolumeClaim.ClaimName)
		}
	}

	var totalCPUReq, totalCPULim, totalMemReq, totalMemLim int64
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				totalCPUReq += cpu.MilliValue()
			}
			if mem := container.Resources.Requests.Memory(); mem != nil {
				totalMemReq += mem.Value()
			}
		}
		if container.Resources.Limits != nil {
			if cpu := container.Resources.Limits.Cpu(); cpu != nil {
				totalCPULim += cpu.MilliValue()
			}
			if mem := container.Resources.Limits.Memory(); mem != nil {
				totalMemLim += mem.Value()
			}
		}
	}

	podResource.CPURequest = formatCPUValue(float64(totalCPUReq))
	podResource.CPULimit = formatCPUValue(float64(totalCPULim))
	podResource.MemoryRequest = formatMemoryValue(float64(totalMemReq) / (1024 * 1024))
	podResource.MemoryLimit = formatMemoryValue(float64(totalMemLim) / (1024 * 1024))
	podResource.CPUActual = "Н/Д"
	podResource.MemoryActual = "Н/Д"

	return podResource
}

func getPodActualUsage(namespace string) map[string]map[string]string {
	ctx := context.Background()
	usage := make(map[string]map[string]string)

	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("⚠️  Предупреждение: не удалось получить метрики подов в %s: %v", namespace, err))
		return usage
	}

	for _, podMetric := range podMetrics.Items {
		var totalCPU, totalMem int64
		for _, container := range podMetric.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMem += container.Usage.Memory().Value()
		}

		usage[podMetric.Name] = map[string]string{
			"cpu":    formatCPUValue(float64(totalCPU)),
			"memory": formatMemoryValue(float64(totalMem) / (1024 * 1024)),
		}
	}

	return usage
}

func getPVCsInfo() []*PVCInfo {
	ctx := context.Background()
	pvcList, err := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("Ошибка получения PVC: %v", err))
		return []*PVCInfo{}
	}

	var pvcs []*PVCInfo
	for _, pvc := range pvcList.Items {
		pvcInfo := &PVCInfo{
			Namespace: pvc.Namespace,
			Name:      pvc.Name,
			Status:    string(pvc.Status.Phase),
			Volume:    pvc.Spec.VolumeName,
		}

		if capacity, ok := pvc.Status.Capacity[corev1.ResourceStorage]; ok {
			pvcInfo.Capacity = capacity.String()
			pvcInfo.Requested = capacity.String()
		}

		if pvc.Spec.StorageClassName != nil {
			pvcInfo.StorageClass = *pvc.Spec.StorageClassName
		}

		for _, mode := range pvc.Spec.AccessModes {
			pvcInfo.AccessModes = append(pvcInfo.AccessModes, string(mode))
		}

		// Фактическое использование недоступно через API — нужен дополнительный мониторинг
		pvcInfo.Used = "0"
		pvcInfo.UsedPercent = 0.0

		pvcs = append(pvcs, pvcInfo)
	}

	return pvcs
}

func getPVsInfo() []*PVInfo {
	ctx := context.Background()
	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		printError(fmt.Sprintf("Ошибка получения PV: %v", err))
		return []*PVInfo{}
	}

	var pvs []*PVInfo
	for _, pv := range pvList.Items {
		pvInfo := &PVInfo{
			Name:   pv.Name,
			Status: string(pv.Status.Phase),
		}

		if capacity, ok := pv.Spec.Capacity[corev1.ResourceStorage]; ok {
			pvInfo.Capacity = capacity.String()
		}

		pvInfo.StorageClass = pv.Spec.StorageClassName

		if pv.Spec.ClaimRef != nil {
			pvInfo.Claim = fmt.Sprintf("%s/%s", pv.Spec.ClaimRef.Namespace, pv.Spec.ClaimRef.Name)
		}

		// Фактическое использование недоступно через API — нужен дополнительный мониторинг
		pvInfo.Used = "0"
		pvInfo.UsedPercent = 0.0

		pvs = append(pvs, pvInfo)
	}

	return pvs
}

func getGatekeeperStatus() *GatekeeperStatus {
	status := &GatekeeperStatus{}
	ctx := context.Background()

	_, err := clientset.CoreV1().Namespaces().Get(ctx, "gatekeeper-system", metav1.GetOptions{})
	if err != nil {
		return status
	}
	status.Installed = true

	pods, err := clientset.CoreV1().Pods("gatekeeper-system").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodRunning {
				status.PodCount++
				status.Running = true
			}
		}
	}

	if dynamicClient == nil {
		return status
	}

	ctGVR := schema.GroupVersionResource{
		Group:    "templates.gatekeeper.sh",
		Version:  "v1",
		Resource: "constrainttemplates",
	}
	ctList, err := dynamicClient.Resource(ctGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		// Пробуем версию v1beta1 как запасной вариант
		ctGVR.Version = "v1beta1"
		ctList, err = dynamicClient.Resource(ctGVR).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return status
	}

	for _, ct := range ctList.Items {
		tmplInfo := ConstraintTemplateInfo{Name: ct.GetName()}
		if spec, ok := ct.Object["spec"].(map[string]interface{}); ok {
			if crd, ok := spec["crd"].(map[string]interface{}); ok {
				if crdSpec, ok := crd["spec"].(map[string]interface{}); ok {
					if names, ok := crdSpec["names"].(map[string]interface{}); ok {
						tmplInfo.Kind, _ = names["kind"].(string)
					}
				}
			}
		}
		status.ConstraintTemplates = append(status.ConstraintTemplates, tmplInfo)
	}

	// Используем Discovery API, а не угадываем имя ресурса из Kind (pluralization нестандартна)
	constraintResList, err := clientset.Discovery().ServerResourcesForGroupVersion("constraints.gatekeeper.sh/v1beta1")
	if err != nil {
		return status
	}
	for _, res := range constraintResList.APIResources {
		if strings.Contains(res.Name, "/") {
			continue // субресурсы пропускаем
		}
		constraintGVR := schema.GroupVersionResource{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: res.Name,
		}
		cList, err := dynamicClient.Resource(constraintGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}
		for _, c := range cList.Items {
			cInfo := ConstraintInfo{
				Name: c.GetName(),
				Kind: c.GetKind(),
			}
			if spec, ok := c.Object["spec"].(map[string]interface{}); ok {
				cInfo.EnforcementAction, _ = spec["enforcementAction"].(string)
				if cInfo.EnforcementAction == "" {
					cInfo.EnforcementAction = "deny"
				}
				if match, ok := spec["match"].(map[string]interface{}); ok {
					if nsList, ok := match["namespaces"].([]interface{}); ok {
						for _, ns := range nsList {
							if nsStr, ok := ns.(string); ok {
								cInfo.Namespaces = append(cInfo.Namespaces, nsStr)
							}
						}
					}
				}
			}
			if statusObj, ok := c.Object["status"].(map[string]interface{}); ok {
				switch v := statusObj["totalViolations"].(type) {
				case int64:
					cInfo.TotalViolations = int(v)
				case float64:
					cInfo.TotalViolations = int(v)
				}
			}
			status.Constraints = append(status.Constraints, cInfo)
		}
	}

	return status
}

func getRBACEntries() []*RBACEntry {
	ctx := context.Background()
	var entries []*RBACEntry

	crbs, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, crb := range crbs.Items {
			for _, subject := range crb.Subjects {
				entry := &RBACEntry{
					Subject:     subject.Name,
					SubjectKind: subject.Kind,
					SubjectNS:   subject.Namespace,
					Role:        crb.RoleRef.Name,
					RoleKind:    crb.RoleRef.Kind,
					BindingName: crb.Name,
					BindingKind: "ClusterRoleBinding",
					Scope:       "cluster",
					BoundIn:     "",
				}
				entries = append(entries, entry)
			}
		}
	}

	rbs, err := clientset.RbacV1().RoleBindings("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, rb := range rbs.Items {
			for _, subject := range rb.Subjects {
				entry := &RBACEntry{
					Subject:     subject.Name,
					SubjectKind: subject.Kind,
					SubjectNS:   subject.Namespace,
					Role:        rb.RoleRef.Name,
					RoleKind:    rb.RoleRef.Kind,
					BindingName: rb.Name,
					BindingKind: "RoleBinding",
					Scope:       "namespace",
					BoundIn:     rb.Namespace,
				}
				entries = append(entries, entry)
			}
		}
	}

	return entries
}
