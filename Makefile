# K8s Resource Analyzer - Makefile
# Кросс-компиляция для Windows, Linux, macOS

BINARY_NAME=k8s-resource-analyzer
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "v8.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w"

GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m

.PHONY: all build clean install help build-all

all: build

## help: Показать справку
help:
	@echo "╔════════════════════════════════════════════════════════════╗"
	@echo "║       K8s Resource Analyzer - Команды сборки               ║"
	@echo "╚════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "$(BLUE)Основные команды:$(NC)"
	@echo "  $(GREEN)make install$(NC)    - Установить зависимости"
	@echo "  $(GREEN)make build$(NC)      - Собрать для текущей платформы"
	@echo "  $(GREEN)make build-all$(NC)  - Собрать для всех платформ"
	@echo "  $(GREEN)make clean$(NC)      - Очистить собранные файлы"
	@echo "  $(GREEN)make run$(NC)        - Запустить приложение"
	@echo ""
	@echo "$(BLUE)Результаты сборки:$(NC)"
	@echo "  Бинарники будут в директории $(YELLOW)bin/$(NC)"
	@echo ""

## install: Установить зависимости
install:
	@echo "$(GREEN)📦 Установка зависимостей...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Зависимости установлены$(NC)"

## build: Собрать для текущей платформы
build:
	@echo "$(GREEN)🔨 Сборка $(BINARY_NAME) для текущей платформы...$(NC)"
	@mkdir -p bin
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/k8s-analyzer
	@echo "$(GREEN)✅ Готово: bin/$(BINARY_NAME)$(NC)"
	@ls -lh bin/$(BINARY_NAME)

## build-all: Собрать для всех платформ (Windows, Linux, macOS)
build-all: clean install
	@echo "$(GREEN)╔════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║  Сборка для всех платформ              ║$(NC)"
	@echo "$(GREEN)╚════════════════════════════════════════╝$(NC)"
	@mkdir -p bin
	@echo ""
	@echo "$(YELLOW)→ Linux AMD64...$(NC)"
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/k8s-analyzer
	@echo "$(GREEN)  ✓ bin/$(BINARY_NAME)-linux-amd64$(NC)"
	@echo ""
	@echo "$(YELLOW)→ Linux ARM64...$(NC)"
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/k8s-analyzer
	@echo "$(GREEN)  ✓ bin/$(BINARY_NAME)-linux-arm64$(NC)"
	@echo ""
	@echo "$(YELLOW)→ macOS Intel (x86_64)...$(NC)"
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/k8s-analyzer
	@echo "$(GREEN)  ✓ bin/$(BINARY_NAME)-darwin-amd64$(NC)"
	@echo ""
	@echo "$(YELLOW)→ macOS Apple Silicon (ARM64)...$(NC)"
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/k8s-analyzer
	@echo "$(GREEN)  ✓ bin/$(BINARY_NAME)-darwin-arm64$(NC)"
	@echo ""
	@echo "$(YELLOW)→ Windows AMD64...$(NC)"
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/k8s-analyzer
	@echo "$(GREEN)  ✓ bin/$(BINARY_NAME)-windows-amd64.exe$(NC)"
	@echo ""
	@echo "$(GREEN)╔════════════════════════════════════════╗$(NC)"
	@echo "$(GREEN)║  ✅ Сборка завершена!                  ║$(NC)"
	@echo "$(GREEN)╚════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(BLUE)Созданные бинарники:$(NC)"
	@ls -lh bin/

## clean: Очистить собранные файлы
clean:
	@echo "$(GREEN)🧹 Очистка...$(NC)"
	@rm -rf bin/
	@go clean
	@echo "$(GREEN)✅ Очистка завершена$(NC)"

## run: Запустить приложение
run:
	@echo "$(GREEN)🚀 Запуск k8s-analyzer...$(NC)"
	@go run ./cmd/k8s-analyzer

## run-buffer: Запустить с кастомным процентом запаса
run-buffer:
	@echo "$(GREEN)🚀 Запуск k8s-analyzer с запасом 30%...$(NC)"
	@go run ./cmd/k8s-analyzer -b 30
