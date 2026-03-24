package main

// Пороги эффективности использования ресурсов
const (
	EfficiencyCritical = 100.0 // Критический уровень (перегруз)
	EfficiencyHigh     = 80.0  // Высокий уровень использования
	EfficiencyNormal   = 50.0  // Нормальный уровень
	EfficiencyLow      = 30.0  // Низкий уровень (недогруз)
)

// Множители для расчетов
const (
	MillicoresInCore = 1000.0 // Миллиядер в одном ядре CPU
	MiBInGiB         = 1024.0 // Мебибайт в гибибайте
)

const (
	ConsoleWidth = 75
)

const (
	SafetyMarginUnderutilized = 1.3 // +30% для недогруженных подов
	SafetyMarginNormal        = 1.2 // +20% для нормально загруженных подов
)

// Версия приложения — перезаписывается GoReleaser через -ldflags "-X main.AppVersion=..."
// При сборке без GoReleaser (go build вручную) остаётся "dev"
var AppVersion = "dev"

var bufferPercent int
