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
pr-reviewer-service/
├── cmd/
│   └── server/
│       └── main.go                   # Точка входа, инициализация зависимостей и HTTP-роуты
├── internal/
│   ├── handlers/
│   │   ├── common.go                 # Общие функции для HTTP-ответов и ошибок
│   │   └── teams.go                  # Хендлеры для работы с командами
│   ├── models/
│   │   └── models.go                 # Доменные модели и структуры API-ответов
│   ├── service/
│   │   └── service.go                # Бизнес-логика назначения и переназначения ревьюеров
│   └── storage/
│       └── postgres.go               # SQL-запросы и работа с PostgreSQL
├── migrations/
│   └── 001_init.sql                  # Начальная миграция: таблицы teams/users/pull_requests
├── результаты тестирования/          # Скриншоты результатов нагрузочного теста
├── Dockerfile                        # Docker-образ приложения
├── docker-compose.yml                # Локальный запуск приложения вместе с PostgreSQL
├── Makefile                          # Команды сборки, запуска и вспомогательные цели
├── go.mod                            # Go-модуль и версия языка
├── go.sum                            # Контрольные суммы зависимостей
└── load-test.bat                     # Нагрузочный тест (Windows)
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
