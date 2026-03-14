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

Проверка состояния:

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

### `POST /team/add`
Создаёт новую команду и пользователей из массива `members`.

**Тело запроса:**
```json
{
  "team_name": "backend",
  "members": [
    {"user_id": "u1", "username": "alice", "is_active": true},
    {"user_id": "u2", "username": "bob", "is_active": true}
  ]
}
```

### `GET /team/get?team_name=...`
Возвращает данные команды по имени (`team_name`) и список её участников.

### `POST /users/setIsActive`
Меняет флаг активности пользователя (`is_active`).

**Тело запроса:**
```json
{
  "user_id": "u2",
  "is_active": false
}
```

### `GET /users/getReview?user_id=...`
Возвращает список PR, где указанный пользователь назначен ревьюером.

### `POST /pullRequest/create`
Создаёт PR и автоматически назначает до двух активных ревьюеров из команды автора.

**Тело запроса:**
```json
{
  "pull_request_id": "pr-101",
  "pull_request_name": "Add auth middleware",
  "author_id": "u1"
}
```

### `POST /pullRequest/reassign`
Переназначает конкретного ревьюера PR на другого активного участника команды.

**Тело запроса:**
```json
{
  "pull_request_id": "pr-101",
  "old_user_id": "u2"
}
```

### `POST /pullRequest/merge`
Переводит PR в статус `MERGED`. После этого переназначение ревьюеров для этого PR запрещено.

**Тело запроса:**
```json
{
  "pull_request_id": "pr-101"
}
```

### `GET /stats`
Возвращает агрегированную статистику по сервису:
- количество команд, пользователей и PR;
- количество открытых и смерженных PR;
- топ пользователей по количеству назначений ревьюером.

### `GET /health`
Проверка доступности сервиса. Возвращает `200 OK` и тело `OK`.

## Полезные команды

- `make help` — выводит все доступные цели Makefile и короткое описание.
- `make build` — собирает бинарник сервиса в `bin/pr-reviewer-service`.
- `make run` — собирает и запускает сервис локально (без Docker).
- `make docker-up` — поднимает приложение и PostgreSQL через Docker Compose (`--build`).
- `make docker-down` — останавливает и удаляет контейнеры Docker Compose.
- `make clean` — удаляет каталог `bin` и очищает кэш сборки Go.
- `make stats` — делает запрос к `http://localhost:8080/stats` и печатает статистику.
- `make load-test` — запускает `load-test.bat` (актуально для Windows).

## Дополнительно

В репозитории есть результаты нагрузочного тестирования в папке `результаты тестирования/`.
