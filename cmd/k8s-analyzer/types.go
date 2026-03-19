package main

// ClusterSummary - сводка по всему кластеру
type ClusterSummary struct {
	TotalPods               int                          // Общее количество подов
	TotalCPURequest         float64                      // Общий CPU request
	TotalCPUActual          float64                      // Общее фактическое использование CPU
	TotalCPURecommended     float64                      // Рекомендуемый CPU (от requests)
	TotalCPUOptimized       float64                      // Оптимизированный CPU (от actual)
	TotalMemRequest         float64                      // Общая память request
	TotalMemActual          float64                      // Общее фактическое использование памяти
	TotalMemRecommended     float64                      // Рекомендуемая память (от requests)
	TotalMemOptimized       float64                      // Оптимизированная память (от actual)
	TotalNodes              int                          // Количество нод
	TotalNodeCPUCapacity    float64                      // Общая CPU емкость нод
	TotalNodeMemoryCapacity float64                      // Общая память емкость нод
	TotalPVCs               int                          // Количество PVC
	TotalPVCCapacity        float64                      // Общая емкость PVC
	TotalPVCUsed            float64                      // Использовано PVC
	TotalPVs                int                          // Количество PV
	TotalPVCapacity         float64                      // Общая емкость PV
	TotalPVUsed             float64                      // Использовано PV
	MaxPodCPUActual         float64                      // Максимальное использование CPU подом
	MaxPodMemoryActual      float64                      // Максимальное использование памяти подом
	MaxPodCPURequest        float64                      // Максимальный CPU request
	MaxPodMemoryRequest     float64                      // Максимальный memory request
	MaxPodNameCPU           string                       // Имя пода с максимальным CPU
	MaxPodNameMemory        string                       // Имя пода с максимальной памятью
	MaxPodNamespaceCPU      string                       // Неймспейс пода с максимальным CPU
	MaxPodNamespaceMemory   string                       // Неймспейс пода с максимальной памятью
	ByNamespace             map[string]*NamespaceSummary // Сводка по неймспейсам
	ByNode                  map[string]*NodeInfo         // Информация по нодам
	ByPVC                   map[string]*PVCInfo          // Информация по PVC
	ByPV                    map[string]*PVInfo           // Информация по PV
}

// NamespaceSummary - сводка по неймспейсу
type NamespaceSummary struct {
	PodCount            int     // Количество подов
	CPURequestTotal     float64 // Общий CPU request
	CPUActualTotal      float64 // Общее фактическое использование CPU
	CPURecommendedTotal float64 // Рекомендуемый CPU
	MemRequestTotal     float64 // Общая память request
	MemActualTotal      float64 // Общее фактическое использование памяти
	MemRecommendedTotal float64 // Рекомендуемая память
}

// NodeInfo - информация о ноде кластера
type NodeInfo struct {
	Name              string  // Имя ноды
	CPUCapacity       float64 // Емкость CPU (в millicores)
	MemoryCapacity    float64 // Емкость памяти (в MiB)
	CPURequests       float64 // Запрошенный CPU (requests)
	MemoryRequests    float64 // Запрошенная память (requests)
	CPUActual         float64 // Фактическое использование CPU
	MemoryActual      float64 // Фактическое использование памяти
	CPUUtilization    float64 // Процент использования CPU (факт)
	MemoryUtilization float64 // Процент использования памяти (факт)
	CPURequestUtil    float64 // Процент от емкости (requests)
	MemoryRequestUtil float64 // Процент от емкости (requests)
	PodsCount         int     // Количество подов на ноде
	Recommendation    string  // Рекомендации по оптимизации
}

// PodResource - информация о ресурсах пода
type PodResource struct {
	Namespace      string   // Неймспейс
	Name           string   // Имя пода
	NodeName       string   // Нода, на которой запущен под
	CPURequest     string   // CPU request
	CPULimit       string   // CPU limit
	CPUActual      string   // Фактическое использование CPU
	MemoryRequest  string   // Memory request
	MemoryLimit    string   // Memory limit
	MemoryActual   string   // Фактическое использование памяти
	RecommendedCPU string   // Рекомендуемый CPU
	RecommendedMem string   // Рекомендуемая память
	Recommendation string   // Рекомендации
	Status         string   // Статус эффективности
	PVCs           []string // Список привязанных PVC
}

// PVCInfo - информация о Persistent Volume Claim
type PVCInfo struct {
	Namespace    string   // Неймспейс
	Name         string   // Имя PVC
	Status       string   // Статус (Bound/Pending/etc)
	Volume       string   // Привязанный PV
	Capacity     string   // Емкость
	Requested    string   // Запрошенный размер
	Used         string   // Использовано
	UsedPercent  float64  // Процент использования
	StorageClass string   // Класс хранилища
	AccessModes  []string // Режимы доступа
}

// PVInfo - информация о Persistent Volume
type PVInfo struct {
	Name         string  // Имя PV
	Capacity     string  // Емкость
	Used         string  // Использовано
	UsedPercent  float64 // Процент использования
	Status       string  // Статус
	Claim        string  // Привязанный PVC
	StorageClass string  // Класс хранилища
}
