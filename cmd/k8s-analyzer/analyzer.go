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

// generatePodRecommendation - генерация рекомендаций для пода.
//
// Структура вывода:
//  1. Сводная строка: факт / рекомендуется / действие по лимиту
//  2. Детальные рекомендации по памяти и CPU
func generatePodRecommendation(pod *PodResource, memEff, cpuEff float64) string {
	var rec strings.Builder

	memAct := parseMemoryValue(pod.MemoryActual)
	memLim := parseMemoryValue(pod.MemoryLimit)
	memReq := parseMemoryValue(pod.MemoryRequest)
	cpuAct := parseCPUValue(pod.CPUActual)
	cpuLim := parseCPUValue(pod.CPULimit)
	cpuReq := parseCPUValue(pod.CPURequest)

	bufferMul := 1.0 + float64(bufferPercent)/100.0

	// ── Сводная строка CPU ─────────────────────────────────────────────────────
	if cpuAct > 0 {
		recLim := cpuAct * bufferMul
		cpuLine := fmt.Sprintf("CPU  | факт: %s | рек. limit (факт+%d%%): %s",
			formatCPUValue(cpuAct), bufferPercent, formatCPUValue(recLim))
		if cpuLim > 0 {
			if cpuLim > recLim*1.1 {
				cpuLine += fmt.Sprintf(" | текущий limit %s → можно снизить на %s",
					formatCPUValue(cpuLim), formatCPUValue(cpuLim-recLim))
			} else if cpuAct >= cpuLim*0.9 {
				cpuLine += fmt.Sprintf(" | текущий limit %s → нужно повысить!", formatCPUValue(cpuLim))
			} else {
				cpuLine += fmt.Sprintf(" | текущий limit %s → ок", formatCPUValue(cpuLim))
			}
		} else {
			cpuLine += " | limit не задан"
		}
		rec.WriteString(cpuLine + "\n")
	}

	// ── Сводная строка Memory ──────────────────────────────────────────────────
	if memAct > 0 {
		recLim := memAct * bufferMul
		memLine := fmt.Sprintf("MEM  | факт: %s | рек. limit (факт+%d%%): %s",
			formatMemoryValue(memAct), bufferPercent, formatMemoryValue(recLim))
		if memLim > 0 {
			if memLim > recLim*1.1 {
				memLine += fmt.Sprintf(" | текущий limit %s → можно снизить на %s",
					formatMemoryValue(memLim), formatMemoryValue(memLim-recLim))
			} else if memAct >= memLim*0.9 {
				memLine += fmt.Sprintf(" | текущий limit %s → нужно повысить!", formatMemoryValue(memLim))
			} else {
				memLine += fmt.Sprintf(" | текущий limit %s → ок", formatMemoryValue(memLim))
			}
		} else {
			memLine += " | limit не задан"
		}
		rec.WriteString(memLine + "\n")
	}

	// ── Детальные предупреждения ───────────────────────────────────────────────
	analyzeMemoryRecommendations(&rec, memEff, memReq, memAct, memLim, pod.MemoryRequest)
	analyzeCPURecommendations(&rec, cpuEff, cpuReq, cpuAct, cpuLim, pod.CPURequest)

	result := strings.TrimSpace(rec.String())
	if result == "" {
		return "✅ Оптимально"
	}
	return result
}

// analyzeMemoryRecommendations - анализ рекомендаций по памяти.
//
// Использует глобальный bufferPercent как целевой запас:
//   рекомендуемый лимит = факт × (1 + bufferPercent/100)
//
// Порядок приоритетов:
//  1. Оценка лимита (влияет на OOMKill)
//  2. Оценка request (влияет на планировщик)
func analyzeMemoryRecommendations(rec *strings.Builder, memEff, memReq, memAct, memLim float64, memReqStr string) {
	bufferMul := 1.0 + float64(bufferPercent)/100.0 // например 1.30 при -b 30
	recommendedLim := memAct * bufferMul

	// ── Анализ лимита (главное — OOMKill) ─────────────────────────────────────

	if memLim <= 0 {
		rec.WriteString("🔴 Memory limit не задан — под может занять всю свободную память ноды\n")
	} else if memAct > 0 {
		limitRatio := (memAct / memLim) * 100.0
		targetRatio := (1.0 / bufferMul) * 100.0 // при буфере 30% = 76.9% — порог "близко к лимиту"

		switch {
		case limitRatio >= 90.0:
			// Критично: потребление вплотную к лимиту, OOMKill неизбежен
			rec.WriteString(fmt.Sprintf(
				"🚨 OOMKill риск: факт %s = %.0f%% от limit %s (буфер %d%% → нужен limit ≥ %s)\n",
				formatMemoryValue(memAct), limitRatio, formatMemoryValue(memLim),
				bufferPercent, formatMemoryValue(recommendedLim)))
		case limitRatio >= targetRatio:
			// Запас меньше заданного буфера — при всплеске нагрузки OOMKill возможен
			rec.WriteString(fmt.Sprintf(
				"⚠️ Запас памяти ниже буфера %d%%: факт %s = %.0f%% от limit %s — рекомендуется limit: %s\n",
				bufferPercent, formatMemoryValue(memAct), limitRatio,
				formatMemoryValue(memLim), formatMemoryValue(recommendedLim)))
		case memLim > recommendedLim*1.5:
			// Лимит завышен более чем в 1.5× от рекомендуемого с учётом буфера
			savings := memLim - recommendedLim
			rec.WriteString(fmt.Sprintf(
				"💡 Limit памяти можно снизить: факт %s × буфер %d%% = %s (текущий %s, экономия %s)\n",
				formatMemoryValue(memAct), bufferPercent,
				formatMemoryValue(recommendedLim), formatMemoryValue(memLim),
				formatMemoryValue(savings)))
		}
	}

	// ── Анализ request (влияет на планировщик, не на OOMKill) ─────────────────

	if memReq > 0 && memAct > 0 {
		recommendedReq := memAct * bufferMul
		switch {
		case memReq > memAct*2:
			factor := memReq / memAct
			rec.WriteString(fmt.Sprintf(
				"⚠️ Memory request завышен в %.1fx (request %s, факт %s) — планировщик занимает лишнее место на ноде; рекомендуется: %s\n",
				factor, memReqStr, formatMemoryValue(memAct), formatMemoryValue(recommendedReq)))
		case memReq < memAct*0.5:
			factor := memAct / memReq
			rec.WriteString(fmt.Sprintf(
				"⚠️ Memory request занижен в %.1fx (request %s, факт %s) — планировщик может перегрузить ноду; рекомендуется: %s\n",
				factor, memReqStr, formatMemoryValue(memAct), formatMemoryValue(recommendedReq)))
		}
	}

	// Request не должен превышать limit
	if memLim > 0 && memReq > memLim {
		rec.WriteString(fmt.Sprintf(
			"🔴 Невалидная конфигурация: memory request (%s) > limit (%s) — под никогда не запустится\n",
			memReqStr, formatMemoryValue(memLim)))
	}
}

// analyzeCPURecommendations - анализ рекомендаций по CPU.
//
// Использует глобальный bufferPercent как целевой запас:
//   рекомендуемый лимит = факт × (1 + bufferPercent/100)
//
// Порядок приоритетов:
//  1. Оценка лимита (влияет на throttling)
//  2. Оценка request (влияет на планировщик)
func analyzeCPURecommendations(rec *strings.Builder, cpuEff, cpuReq, cpuAct, cpuLim float64, cpuReqStr string) {
	bufferMul := 1.0 + float64(bufferPercent)/100.0
	recommendedLim := cpuAct * bufferMul

	nl := func() {
		if rec.Len() > 0 {
			rec.WriteString("\n")
		}
	}

	// ── Анализ лимита (throttling) ─────────────────────────────────────────────

	if cpuLim <= 0 {
		nl()
		rec.WriteString("💡 CPU limit не задан — под может использовать всё свободное CPU ноды (допустимо, но следите за соседями)")
	} else if cpuAct > 0 {
		limitRatio := (cpuAct / cpuLim) * 100.0
		targetRatio := (1.0 / bufferMul) * 100.0

		switch {
		case limitRatio >= 90.0:
			nl()
			rec.WriteString(fmt.Sprintf(
				"🚨 CPU throttling: факт %s = %.0f%% от limit %s (буфер %d%% → нужен limit ≥ %s)",
				formatCPUValue(cpuAct), limitRatio, formatCPUValue(cpuLim),
				bufferPercent, formatCPUValue(recommendedLim)))
		case limitRatio >= targetRatio:
			nl()
			rec.WriteString(fmt.Sprintf(
				"⚠️ Запас CPU ниже буфера %d%%: факт %s = %.0f%% от limit %s — рекомендуется limit: %s",
				bufferPercent, formatCPUValue(cpuAct), limitRatio,
				formatCPUValue(cpuLim), formatCPUValue(recommendedLim)))
		case cpuLim > recommendedLim*1.5:
			savings := cpuLim - recommendedLim
			nl()
			rec.WriteString(fmt.Sprintf(
				"💡 CPU limit можно снизить: факт %s × буфер %d%% = %s (текущий %s, экономия %s)",
				formatCPUValue(cpuAct), bufferPercent,
				formatCPUValue(recommendedLim), formatCPUValue(cpuLim),
				formatCPUValue(savings)))
		}
	}

	// ── Анализ request ─────────────────────────────────────────────────────────

	if cpuReq > 0 && cpuAct > 0 {
		recommendedReq := cpuAct * bufferMul
		switch {
		case cpuReq > cpuAct*2:
			factor := cpuReq / cpuAct
			nl()
			rec.WriteString(fmt.Sprintf(
				"⚠️ CPU request завышен в %.1fx (request %s, факт %s) — рекомендуется: %s",
				factor, cpuReqStr, formatCPUValue(cpuAct), formatCPUValue(recommendedReq)))
		case cpuReq < cpuAct*0.5:
			factor := cpuAct / cpuReq
			nl()
			rec.WriteString(fmt.Sprintf(
				"⚠️ CPU request занижен в %.1fx (request %s, факт %s) — рекомендуется: %s",
				factor, cpuReqStr, formatCPUValue(cpuAct), formatCPUValue(recommendedReq)))
		}
	}

	// Request не должен превышать limit
	if cpuLim > 0 && cpuReq > cpuLim {
		nl()
		rec.WriteString(fmt.Sprintf(
			"🔴 Невалидная конфигурация: CPU request (%s) > limit (%s) — под никогда не запустится",
			cpuReqStr, formatCPUValue(cpuLim)))
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

// calculateRecommendedLimit - рекомендуемый лимит = факт × (1 + bufferPercent/100).
// isMemory=true форматирует как память, false — как CPU.
func calculateRecommendedLimit(actualStr string, isMemory bool) string {
	bufferMul := 1.0 + float64(bufferPercent)/100.0
	if isMemory {
		act := parseMemoryValue(actualStr)
		if act <= 0 {
			return ""
		}
		return formatMemoryValue(act * bufferMul)
	}
	act := parseCPUValue(actualStr)
	if act <= 0 {
		return ""
	}
	return formatCPUValue(act * bufferMul)
}
