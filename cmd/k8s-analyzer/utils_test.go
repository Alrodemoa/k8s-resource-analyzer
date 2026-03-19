package main

import (
	"testing"
)

// ============================================================================
// Парсинг памяти
// ============================================================================

func TestParseMemoryValue_GiSuffix(t *testing.T) {
	got := parseMemoryValue("2Gi")
	want := 2.0 * MiBInGiB
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestParseMemoryValue_MiSuffix(t *testing.T) {
	got := parseMemoryValue("512Mi")
	if got != 512.0 {
		t.Errorf("получено %.2f, ожидалось 512.0", got)
	}
}

func TestParseMemoryValue_KiSuffix(t *testing.T) {
	got := parseMemoryValue("1024Ki")
	want := 1024.0 / MiBInGiB
	if got != want {
		t.Errorf("получено %.4f, ожидалось %.4f", got, want)
	}
}

func TestParseMemoryValue_GSuffix(t *testing.T) {
	got := parseMemoryValue("1g")
	want := 1.0 * MiBInGiB
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestParseMemoryValue_CaseInsensitive(t *testing.T) {
	a := parseMemoryValue("1GI")
	b := parseMemoryValue("1gi")
	if a != b {
		t.Errorf("несоответствие без учёта регистра: %.2f != %.2f", a, b)
	}
}

func TestParseMemoryValue_NA(t *testing.T) {
	for _, v := range []string{"n/a", "N/A", "", "<none>", "Н/Д"} {
		got := parseMemoryValue(v)
		if got != 0 {
			t.Errorf("parseMemoryValue(%q) = %.2f, ожидалось 0", v, got)
		}
	}
}

func TestParseMemoryValue_Plain(t *testing.T) {
	got := parseMemoryValue("1024")
	if got != 1024.0 {
		t.Errorf("получено %.2f, ожидалось 1024.0", got)
	}
}

// ============================================================================
// Парсинг CPU
// ============================================================================

func TestParseCPUValue_Millicores(t *testing.T) {
	got := parseCPUValue("500m")
	if got != 500.0 {
		t.Errorf("получено %.2f, ожидалось 500.0", got)
	}
}

func TestParseCPUValue_Cores(t *testing.T) {
	got := parseCPUValue("2")
	want := 2.0 * MillicoresInCore
	if got != want {
		t.Errorf("получено %.2f, ожидалось %.2f", got, want)
	}
}

func TestParseCPUValue_NA(t *testing.T) {
	for _, v := range []string{"n/a", "", "<none>", "Н/Д"} {
		got := parseCPUValue(v)
		if got != 0 {
			t.Errorf("parseCPUValue(%q) = %.2f, ожидалось 0", v, got)
		}
	}
}

func TestParseCPUValue_CaseInsensitive(t *testing.T) {
	a := parseCPUValue("500M")
	b := parseCPUValue("500m")
	if a != b {
		t.Errorf("несоответствие без учёта регистра: %.2f != %.2f", a, b)
	}
}

func TestParseCPUValue_Zero(t *testing.T) {
	got := parseCPUValue("0m")
	if got != 0 {
		t.Errorf("получено %.2f, ожидалось 0", got)
	}
}

// ============================================================================
// Форматирование CPU
// ============================================================================

func TestFormatCPUValue_Normal(t *testing.T) {
	got := formatCPUValue(250)
	if got != "250m" {
		t.Errorf("получено %q, ожидалось \"250m\"", got)
	}
}

func TestFormatCPUValue_Zero(t *testing.T) {
	got := formatCPUValue(0)
	if got != "Н/Д" {
		t.Errorf("получено %q, ожидалось \"Н/Д\"", got)
	}
}

func TestFormatCPUValue_Negative(t *testing.T) {
	got := formatCPUValue(-100)
	if got != "Н/Д" {
		t.Errorf("получено %q, ожидалось \"Н/Д\"", got)
	}
}

func TestFormatCPUValue_RoundTrip(t *testing.T) {
	original := 750.0
	formatted := formatCPUValue(original)
	parsed := parseCPUValue(formatted)
	if parsed != original {
		t.Errorf("обратное преобразование не совпало: %.2f → %q → %.2f", original, formatted, parsed)
	}
}

// ============================================================================
// Форматирование памяти
// ============================================================================

func TestFormatMemoryValue_GiRange(t *testing.T) {
	got := formatMemoryValue(2048) // 2 ГиБ
	if got != "2.0Gi" {
		t.Errorf("получено %q, ожидалось \"2.0Gi\"", got)
	}
}

func TestFormatMemoryValue_MiRange(t *testing.T) {
	got := formatMemoryValue(512)
	if got != "512Mi" {
		t.Errorf("получено %q, ожидалось \"512Mi\"", got)
	}
}

func TestFormatMemoryValue_ExactlyOneGi(t *testing.T) {
	got := formatMemoryValue(MiBInGiB)
	if got != "1.0Gi" {
		t.Errorf("получено %q, ожидалось \"1.0Gi\"", got)
	}
}

func TestFormatMemoryValue_Zero(t *testing.T) {
	got := formatMemoryValue(0)
	if got != "Н/Д" {
		t.Errorf("получено %q, ожидалось \"Н/Д\"", got)
	}
}

func TestFormatMemoryValue_Negative(t *testing.T) {
	got := formatMemoryValue(-256)
	if got != "Н/Д" {
		t.Errorf("получено %q, ожидалось \"Н/Д\"", got)
	}
}

// ============================================================================
// Очистка имени листа
// ============================================================================

func TestSanitizeSheetName_RemovesSpecialChars(t *testing.T) {
	cases := map[string]string{
		"name[0]": "name-0-",
		"a:b":     "a-b",
		"x*y":     "x-y",
		"foo?bar": "foo-bar",
		"a/b":     "a-b",
		`a\b`:     "a-b",
	}
	for input, want := range cases {
		got := sanitizeSheetName(input)
		if got != want {
			t.Errorf("sanitizeSheetName(%q) = %q, ожидалось %q", input, got, want)
		}
	}
}

func TestSanitizeSheetName_TruncatesTo31(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyz123456" // 32 символа
	got := sanitizeSheetName(long)
	if len(got) > 31 {
		t.Errorf("ожидалась длина ≤ 31, получено %d", len(got))
	}
}

func TestSanitizeSheetName_ShortNameUnchanged(t *testing.T) {
	got := sanitizeSheetName("default")
	if got != "default" {
		t.Errorf("получено %q, ожидалось \"default\"", got)
	}
}

// ============================================================================
// Центрирование текста
// ============================================================================

func TestCenterText_Normal(t *testing.T) {
	got := centerText("hi", 10)
	if len(got) < 2 || got[len(got)-2:] != "hi" {
		t.Errorf("текст не найден в результате: %q", got)
	}
	// Должен иметь отступ слева
	if got[0] != ' ' {
		t.Errorf("ожидался пробел в начале строки, получено %q", got)
	}
}

func TestCenterText_TextLongerThanWidth(t *testing.T) {
	got := centerText("hello world", 5)
	if got != "hello world" {
		t.Errorf("ожидался неизменённый текст, получено %q", got)
	}
}

func TestCenterText_ExactWidth(t *testing.T) {
	got := centerText("hello", 5)
	if got != "hello" {
		t.Errorf("ожидался неизменённый текст при точном совпадении ширины, получено %q", got)
	}
}

// ============================================================================
// Парсинг выбора
// ============================================================================

func TestParseSelection_Single(t *testing.T) {
	got := parseSelection("2", 5)
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("получено %v, ожидалось [2]", got)
	}
}

func TestParseSelection_CommaSeparated(t *testing.T) {
	got := parseSelection("1,3,5", 5)
	if len(got) != 3 {
		t.Errorf("получено %v, ожидалось [1 3 5]", got)
	}
}

func TestParseSelection_Range(t *testing.T) {
	got := parseSelection("1-4", 5)
	if len(got) != 4 {
		t.Errorf("получено %v, ожидалось [1 2 3 4]", got)
	}
	for i, v := range got {
		if v != i+1 {
			t.Errorf("got[%d] = %d, ожидалось %d", i, v, i+1)
		}
	}
}

func TestParseSelection_OutOfBounds(t *testing.T) {
	got := parseSelection("10", 5)
	if len(got) != 0 {
		t.Errorf("ожидался пустой результат при выходе за границы, получено %v", got)
	}
}

func TestParseSelection_Deduplication(t *testing.T) {
	got := parseSelection("1,1,2", 5)
	if len(got) != 2 {
		t.Errorf("ожидалась дедупликация, получено %v", got)
	}
}

func TestParseSelection_ZeroIndex(t *testing.T) {
	got := parseSelection("0", 5)
	if len(got) != 0 {
		t.Errorf("ожидался пустой результат для нулевого индекса, получено %v", got)
	}
}

func TestParseSelection_RangeOutOfBounds(t *testing.T) {
	got := parseSelection("3-10", 5)
	if len(got) != 0 {
		t.Errorf("ожидался пустой результат для диапазона за пределами максимума, получено %v", got)
	}
}
