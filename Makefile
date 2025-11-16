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
	@echo "Running load test..."
	load-test.bat

stats:
	@echo "Getting statistics..."
	@curl -s http://localhost:8080/stats

help:
	@echo "Available commands:"
	@echo "  make docker-up    - Start service"
	@echo "  make docker-down  - Stop service"
	@echo "  make build        - Build application"
	@echo "  make run          - Run locally"
	@echo "  make clean        - Clean project"
	@echo "  make load-test    - Load testing"
	@echo "  make stats        - Show statistics"

default: help