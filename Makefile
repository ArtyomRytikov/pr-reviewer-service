.PHONY: build run docker-up docker-down clean load-test stats help

build:
	go build -o bin/pr-reviewer-service ./cmd/server

run: build
	./bin/pr-reviewer-service

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

clean:
	-del /q bin\* 2>nul
	-rmdir /s /q bin 2>nul
	go clean

load-test:
	@echo "Запуск нагрузочного тестирования..."
	load-test.bat
stats:
	@echo "Получение статистики..."
	@curl -s http://localhost:8080/stats

help:
	@echo "Доступные команды:"
	@echo "  make docker-up    - Запустить сервис"
	@echo "  make docker-down  - Остановить сервис"
	@echo "  make build        - Собрать приложение"
	@echo "  make run          - Запустить локально"
	@echo "  make clean        - Очистить проект"
	@echo "  make load-test    - Нагрузочное тестирование"
	@echo "  make stats        - Показать статистику"

default: help