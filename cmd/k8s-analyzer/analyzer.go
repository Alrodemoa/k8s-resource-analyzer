package main

// Модуль анализа и расчётов эффективности ресурсов

import (
	"fmt"
	"strings"
)

// calculateMemoryEfficiency - расчёт эффективности использования памяти
func calculateMemoryEfficiency(pod *PodResource) float64 {
	memReq := parseMemoryValue(pod.MemoryRequest)
	memAct := parseMemoryValue(pod.MemoryActual)

	if memReq <= 0 || memAct <= 0 {
		return 0
	}

	return (memAct / memReq) * 100.0
}

// calculateCPUEfficiency - расчёт эффективности использования CPU
func calculateCPUEfficiency(pod *PodResource) float64 {
	cpuReq := parseCPUValue(pod.CPURequest)
	cpuAct := parseCPUValue(pod.CPUActual)

	if cpuReq <= 0 || cpuAct <= 0 {
		return 0
	}

	return (cpuAct / cpuReq) * 100.0
}

// determinePodStatus - определение статуса пода по эффективности
func determinePodStatus(memEff, cpuEff float64) string {
	maxEff := memEff
	if cpuEff > maxEff {
		maxEff = cpuEff
	}

	if maxEff >= EfficiencyCritical {
		return "🔴 Критично"
	} else if maxEff >= EfficiencyHigh {
		return "🟠 Высокая загрузка"
	} else if maxEff >= EfficiencyNormal {
		return "🟢 Оптимально"
	} else if maxEff >= EfficiencyLow {
		return "🟡 Недогруз"
	}
	return "⚪ Минимальная загрузка"
}

// generatePodRecommendation - генерация рекомендаций для пода
func generatePodRecommendation(pod *PodResource, memEff, cpuEff float64) string {
	var rec strings.Builder

	memReq := parseMemoryValue(pod.MemoryRequest)
	memAct := parseMemoryValue(pod.MemoryActual)
	memLim := parseMemoryValue(pod.MemoryLimit)
	cpuReq := parseCPUValue(pod.CPURequest)
	cpuAct := parseCPUValue(pod.CPUActual)
	cpuLim := parseCPUValue(pod.CPULimit)

	// Анализ памяти
	analyzeMemoryRecommendations(&rec, memEff, memReq, memAct, memLim, pod.MemoryRequest)

	// Анализ CPU
	analyzeCPURecommendations(&rec, cpuEff, cpuReq, cpuAct, cpuLim, pod.CPURequest)

	result := strings.TrimSpace(rec.String())
	if result == "" {
		return "✅ Оптимально"
	}
	return result
}

// analyzeMemoryRecommendations - анализ рекомендаций по памяти
func analyzeMemoryRecommendations(rec *strings.Builder, memEff, memReq, memAct, memLim float64, memReqStr string) {
	if memEff >= EfficiencyCritical {
		factor := memEff / 100.0
		rec.WriteString(fmt.Sprintf("🔴 Память занижена в %.1f раз! Увеличьте %s → %s\n",
			factor, memReqStr, formatMemoryValue(memAct*SafetyMarginNormal)))
	} else if memEff < EfficiencyLow {
		factor := 100.0 / memEff
		rec.WriteString(fmt.Sprintf("⚠️ Память завышена в %.1f раз! Уменьшите %s → %s\n",
			factor, memReqStr, formatMemoryValue(memAct*SafetyMarginUnderutilized)))
	}

	// Проверка на отсутствие limit
	if memLim <= 0 && memReq > 0 {
		rec.WriteString("💡 Рекомендуется установить limit для памяти\n")
	}

	// Предупреждение если лимит близок к фактическому использованию (риск OOMKill)
	if memLim > 0 && memAct > 0 {
		limitUsageRatio := (memAct / memLim) * 100.0
		if limitUsageRatio >= 85.0 {
			rec.WriteString(fmt.Sprintf("🚨 Лимит памяти почти исчерпан (%.0f%% от лимита) — риск OOMKill!\n",
				limitUsageRatio))
		}
	}
	// Предупреждение если лимит намного больше request
	if memLim > 0 && memReq > 0 && memLim > memReq*5 {
		rec.WriteString("⚠️ Лимит памяти в 5+ раз превышает request — возможна утечка памяти\n")
	}
}

// analyzeCPURecommendations - анализ рекомендаций по CPU
func analyzeCPURecommendations(rec *strings.Builder, cpuEff, cpuReq, cpuAct, cpuLim float64, cpuReqStr string) {
	if cpuEff >= EfficiencyCritical {
		factor := cpuEff / 100.0
		rec.WriteString(fmt.Sprintf("🔴 CPU занижен в %.1f раз! Увеличьте %s → %s",
			factor, cpuReqStr, formatCPUValue(cpuAct*SafetyMarginNormal)))
	} else if cpuEff < EfficiencyLow {
		factor := 100.0 / cpuEff
		rec.WriteString(fmt.Sprintf("⚠️ CPU завышен в %.1f раз! Уменьшите %s → %s",
			factor, cpuReqStr, formatCPUValue(cpuAct*SafetyMarginUnderutilized)))
	}

	// Проверка на отсутствие limit
	if cpuLim <= 0 && cpuReq > 0 {
		if rec.Len() > 0 {
			rec.WriteString("\n")
		}
		rec.WriteString("💡 Рекомендуется установить limit для CPU")
	}

	// Предупреждение если лимит CPU близок к фактическому использованию (риск throttling)
	if cpuLim > 0 && cpuAct > 0 {
		limitUsageRatio := (cpuAct / cpuLim) * 100.0
		if limitUsageRatio >= 85.0 {
			if rec.Len() > 0 {
				rec.WriteString("\n")
			}
			rec.WriteString(fmt.Sprintf("🚨 CPU throttling: используется %.0f%% от лимита CPU!", limitUsageRatio))
		}
	}
}

// calculateRecommendedCPU - расчёт рекомендуемого значения CPU
func calculateRecommendedCPU(pod *PodResource) string {
	cpuReq := parseCPUValue(pod.CPURequest)
	cpuAct := parseCPUValue(pod.CPUActual)

	if cpuReq <= 0 || cpuAct <= 0 {
		return pod.CPURequest
	}

	cpuEff := (cpuAct / cpuReq) * 100.0

	var recommended float64
	if cpuEff >= EfficiencyHigh {
		// Высокая эффективность - добавляем запас от actual
		recommended = cpuAct * SafetyMarginNormal
	} else if cpuEff < EfficiencyLow {
		// Низкая эффективность - уменьшаем request
		recommended = cpuAct * SafetyMarginUnderutilized
	} else {
		// Оптимальная эффективность - добавляем процент запаса
		recommended = cpuReq * (1.0 + float64(bufferPercent)/100.0)
	}

	return formatCPUValue(recommended)
}

// calculateRecommendedMemory - расчёт рекомендуемого значения памяти
func calculateRecommendedMemory(pod *PodResource) string {
	memReq := parseMemoryValue(pod.MemoryRequest)
	memAct := parseMemoryValue(pod.MemoryActual)

	if memReq <= 0 || memAct <= 0 {
		return pod.MemoryRequest
	}

	memEff := (memAct / memReq) * 100.0

	var recommended float64
	if memEff >= EfficiencyHigh {
		// Высокая эффективность - добавляем запас от actual
		recommended = memAct * SafetyMarginNormal
	} else if memEff < EfficiencyLow {
		// Низкая эффективность - уменьшаем request
		recommended = memAct * SafetyMarginUnderutilized
	} else {
		// Оптимальная эффективность - добавляем процент запаса
		recommended = memReq * (1.0 + float64(bufferPercent)/100.0)
	}

	return formatMemoryValue(recommended)
}
