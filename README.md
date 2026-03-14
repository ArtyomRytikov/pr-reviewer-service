# PR Reviewer Service

Микросервис для автоматического назначения ревьюеров на Pull Request внутри команды автора.

## Суть проекта

Сервис реализует простой HTTP API для работы с командами и pull request:

- создание и чтение команд;
- управление активностью пользователей (`is_active`);
- создание PR с автоматическим назначением до 2 активных ревьюеров из команды автора;
- переназначение ревьюера на другого активного участника;
- merge PR (после merge переназначение запрещено);
- получение списка PR, назначенных конкретному пользователю;
- получение общей статистики по системе.

## Стек

- **Go 1.21**
- **PostgreSQL 13**
- **Docker / Docker Compose**

## Структура проекта

```text
.
├── cmd/server/main.go                # Точка входа, HTTP-роутинг и запуск сервера
├── internal/models/models.go         # Модели домена и DTO ответов
├── internal/service/service.go       # Бизнес-логика назначения/переназначения
├── internal/storage/postgres.go      # Работа с PostgreSQL
├── internal/handlers/                # Выделенные хендлеры (частично)
├── migrations/001_init.sql           # Инициализация схемы БД
├── docker-compose.yml                # Локальный запуск приложения + PostgreSQL
├── Dockerfile                        # Сборка контейнера приложения
├── Makefile                          # Команды для разработки
└── load-test.bat                     # Скрипт нагрузочного теста (Windows)
```

## Как запустить после клонирования репозитория

### 1. Клонировать репозиторий

```bash
git clone <URL_РЕПОЗИТОРИЯ>
cd pr-reviewer-service
```

### 2. Запустить через Docker Compose (рекомендуется)

```bash
docker compose up --build
```

Сервис будет доступен по адресу: `http://localhost:8080`.

Проверка здоровья:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ: `OK`.

### 3. Альтернатива через Makefile

```bash
make docker-up
```

Остановка:

```bash
make docker-down
```

## Локальный запуск без Docker

Требуется локально поднятый PostgreSQL и корректные переменные окружения.

```bash
go mod download
go build -o bin/pr-reviewer-service ./cmd/server
./bin/pr-reviewer-service
```

Переменные окружения (со значениями по умолчанию):

- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=password`
- `DB_NAME=pr_reviewer`
- `PORT=8080`

## Основные HTTP endpoints

- `POST /team/add`
- `GET /team/get?team_name=...`
- `POST /users/setIsActive`
- `GET /users/getReview?user_id=...`
- `POST /pullRequest/create`
- `POST /pullRequest/reassign`
- `POST /pullRequest/merge`
- `GET /stats`
- `GET /health`

## Полезные команды

```bash
make help
make build
make run
make stats
```

## Дополнительно

В репозитории есть результаты нагрузочного тестирования в папке `результаты тестирования/`.
