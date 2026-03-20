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

.PHONY: all build clean install help

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
