# Agent Service

Эталонный микросервис с чистой слоистой архитектурой на Go.

## Архитектура

```
cmd/
  ├── agent/           # Точка входа сервиса agent
  └── migrator/        # Точка входа миграцій

internal/
  ├── app/
  │   ├── agent/       # Инициализация приложения agent
  │   └── migrator/    # Инициализация миграцій
  ├── api/
  │   └── agent/       # HTTP handlers и routing
  ├── service/
  │   ├── service.go   # Интерфейсы сервисного слоя
  │   └── agent/       # Реализация сервісів
  ├── storage/
  │   ├── storage.go   # Интерфейсы хранилища
  │   └── agent/       # Реализация PostgreSQL хранилища
  └── config/          # Конфігурація приложения

migration/             # SQL миграції
```

## HTTP endpoints

### GET /health
Возвращает статус здоровья сервиса и записывает вызов в БД.

**Response (200 OK):**
```json
{"status":"success"}
```

### GET /check-manager
Проверяет здоровье сервиса manager и записывает результат в БД.

**Response (200 OK - success):**
```json
{
  "status": "success",
  "manager_status": "success",
  "manager_url": "http://localhost:8081",
  "http_status": 200
}
```

**Response (200 OK - error):**
```json
{
  "status": "error",
  "manager_status": "error",
  "manager_url": "http://localhost:8081",
  "http_status": 500,
  "error": "unexpected HTTP status: 500"
}
```

## Требования

- Go 1.23.4+
- PostgreSQL 12+
- Docker (опционально)

## Конфигурация

### app.toml
Основная конфигурация приложения (не чувствительные данные):

```toml
service_name = "agent"
service_env = "local"
http_port = 8080

[database]
host = "localhost"
port = 5432
name = "agent_db"
sslmode = "disable"

[manager]
url = "http://localhost:8081"
timeout_seconds = 5
```

### Переменные окружения

Чувствительные данные:
```
DB_USER=agent_user
DB_PASSWORD=agent_password
LOG_LEVEL=info
```

Для локальной разработки используйте `.env` файл:
```bash
cp .env.example .env
# Отредактируйте .env с нужными значениями
```

## Запуск

### Через Docker Compose (рекомендуется)

```bash
docker-compose up --build
```

Это запустит:
1. PostgreSQL контейнер
2. Migrator для применения миграций
3. Agent сервис на порту 8080

### Локально (при запущенной БД)

#### 1. Подготовка БД

```bash
# Создайте БД и пользователя (если еще не созданы)
psql -U postgres -c "CREATE USER agent_user WITH PASSWORD 'agent_password';"
psql -U postgres -c "CREATE DATABASE agent_db OWNER agent_user;"
```

#### 2. Применение миграций

```bash
export DB_USER=agent_user
export DB_PASSWORD=agent_password
go run ./cmd/migrator/main.go migration
```

#### 3. Запуск сервиса

```bash
export DB_USER=agent_user
export DB_PASSWORD=agent_password
go run ./cmd/agent/main.go
```

Или прямо:
```bash
export DB_USER=agent_user DB_PASSWORD=agent_password
./agent
```

## Тестирование

### Запуск всех тестов

```bash
go test ./...
```

### С подробным выводом

```bash
go test ./... -v
```

### Запуск конкретного пакета

```bash
go test ./internal/service/agent -v
```

## Примеры запросов

### Health check

```bash
curl -X GET http://localhost:8080/health
```

### Check manager

```bash
curl -X GET http://localhost:8080/check-manager
```

## Сборка

### Сборка бинарников

```bash
go build -o agent ./cmd/agent/main.go
go build -o migrator ./cmd/migrator/main.go
```

### Сборка Docker образа

```bash
docker build -t agent:latest .
```

## БД схема

### health_calls
Таблица записей вызовов /health endpoint:

```
id              SERIAL PRIMARY KEY
called_at       TIMESTAMPTZ NOT NULL
```

### manager_checks
Таблица записей проверок health менеджера:

```
id              SERIAL PRIMARY KEY
checked_at      TIMESTAMPTZ NOT NULL
manager_url     TEXT NOT NULL
status          TEXT NOT NULL ('success' или 'error')
http_status     INT NULL
error_message   TEXT NULL
```

## Особенности кода

- **Чистая архитектура**: Разделение на слои (API → Service → Storage)
- **Интерфейсы**: Зависимости через интерфейсы для тестируемости
- **Логирование**: Используется стандартный slog для логирования
- **Обработка ошибок**: Правильная обработка и логирование ошибок
- **Контекст**: Использование context для управления таймаутами
- **Тесты**: Unit тесты с mock объектами

## Лицензия

MIT
