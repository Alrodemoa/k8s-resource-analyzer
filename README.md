<div align="center">

```
  ██╗  ██╗  █████╗   ██████╗    Анализатор ресурсов
  ██║ ██╔╝ ██╔══██╗ ██╔════╝   ─────────────────────────────────────────
  █████╔╝  ╚█████╔╝ ╚█████╗    v1.1.1  ·  Kubernetes Кластер
  ██╔═██╗  ██╔══██╗  ╚════██╗  Анализ и оптимизация ресурсов
  ██║  ██╗ ╚█████╔╝  ██████╔╝  кластера с Excel отчётом
  ╚═╝  ╚═╝  ╚════╝   ╚═════╝
```

**Анализатор ресурсов Kubernetes — от текущего момента до глубокой истории**

[![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/your-org/k8s-resource-analyzer?style=flat-square)](../../releases)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)

</div>

---

## Что это

Инструмент командной строки для глубокого анализа ресурсов Kubernetes-кластера. Собирает метрики CPU и памяти по узлам, неймспейсам и подам, выдаёт конкретные рекомендации по оптимизации лимитов и запросов, экспортирует всё в Excel-отчёт с графиками и цветовой индикацией.

---

## Возможности

### Источники метрик

| Режим | Флаги | Описание |
|-------|-------|----------|
| **Текущий момент** | *(без флагов)* | Снимок из Metrics Server |
| **Живой сбор** | `-d 30m` | Непрерывный опрос Metrics Server за период, min/avg/max/p95 |
| **Prometheus** | `-p http://…:9090 -d 1h` | Исторические данные из Prometheus или Thanos |
| **Thanos (multi-cluster)** | `-p … -d … --cluster prod-eu` | Фильтрация по кластеру; автоопределение по kubectl-контексту (если имя контекста совпадает с именем кластера в Thanos) |

### Анализ и рекомендации

- 🎯 **Рекомендации по лимитам** — риск OOMKill, CPU throttling, потенциальная экономия
- 📊 **Рекомендации по requests** — завышены / занижены относительно факта
- 🔢 **Настраиваемый буфер** — все пороги считаются через `-b <процент>` (по умолчанию 50%)
- 🔒 **OPA Gatekeeper** — список шаблонов и ограничений с режимами применения
- 👥 **RBAC** — сводная таблица SubjectKind / Role / Namespace по всем привязкам

### Отчёт Excel

| Лист | Содержимое |
|------|------------|
| 📊 Сводка кластера | Общая статистика CPU/памяти, графики, потенциальная экономия |
| 🖥️ Узлы | Ёмкость, фактическая нагрузка, utilization по requests |
| 💽 PVC / PV | Persistent-тома: статус, ёмкость, класс хранилища |
| 📦 `<namespace>` | Детализация по подам — факт, requests, limits, рекомендации |
| 📈 История | min / avg / max / p95 за период (при `-d`) |
| 🔒 Gatekeeper | Шаблоны ограничений и политики |
| 👥 RBAC | Привязки ролей кластера и неймспейсов |

---

## Быстрый старт

Скачайте бинарник для вашей платформы со страницы [Releases](../../releases).

```bash
# Снимок текущего момента (буфер 50%)
./k8s-analyzer

# Настроить буфер запаса для расчёта рекомендаций
./k8s-analyzer -b 30    # dev-окружение
./k8s-analyzer -b 100   # prod-окружение

# Живой сбор за 30 минут (Metrics Server)
./k8s-analyzer -d 30m

# История из Prometheus за последний час
./k8s-analyzer -p http://prometheus:9090 -d 1h

# История из Thanos с автоопределением кластера
./k8s-analyzer -p http://thanos:9090 -d 7d

# Явно указать кластер в Thanos
./k8s-analyzer -p http://thanos:9090 -d 1d --cluster prod-eu

# Версия / справка
./k8s-analyzer -v
./k8s-analyzer --help
```

После запуска инструмент предложит выбрать неймспейсы:

```
  1.  default                     7.  monitoring
  2.  kube-system                 8.  ingress-nginx
  3.  production                  ...

  Форматы выбора:
    1,3,5   — через запятую
    1-10    — диапазон
    *       — все неймспейсы

  Ваш выбор: *
```

Результат сохраняется в файл `k8s-анализ-YYYY-MM-DD-HHmmss.xlsx`.

---

## Все опции

```
  -b, --buffer <число>          Процент запаса ресурсов (по умолчанию 50)
  -p, --prometheus <url>        URL Prometheus/Thanos
  -d, --duration <период>       Период: 30m · 2h · 7d · 1w · 2w3d
      --cluster <имя>           Имя кластера в Thanos — если автоопределение не сработало
                                (автоопределение: имя kubectl-контекста сопоставляется со
                                значениями лейбла cluster в Thanos)
      --cluster-label <лейбл>   Лейбл кластера в Thanos, если не 'cluster'
                                (например: region, env, dc)
  -v, --version                 Показать версию
  -h, --help                    Справка
```

---

## Требования

- Настроенный `kubeconfig` (`~/.kube/config` или переменная `KUBECONFIG`)
- [Metrics Server](https://github.com/kubernetes-sigs/metrics-server) в кластере *(не нужен в Prometheus-режиме)*

---

## Сборка из исходников

Требуется Go 1.19+.

**Linux / macOS:**
```bash
go mod download
go build -o k8s-analyzer ./cmd/k8s-analyzer
```

**Windows:**
```cmd
go mod download
go build -o k8s-analyzer.exe .\cmd\k8s-analyzer
```

**Через Makefile (Linux / macOS):**
```bash
make build        # текущая платформа → bin/k8s-resource-analyzer
make build-all    # все платформы → bin/
make run          # go run без сборки
```

---

## Лицензия

[MIT](LICENSE)
