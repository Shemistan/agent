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
Проверяет здоровье всех настроенных сервисов manager и записывает результаты в БД.

**Response (200 OK - все успешно):**
```json
{
  "status": "success",
  "managers": [
    {
      "manager_url": "http://localhost:8081",
      "status": "success",
      "http_status": 200
    },
    {
      "manager_url": "https://manager2:8443",
      "status": "success",
      "http_status": 200
    }
  ]
}
```

**Response (200 OK - одна ошибка):**
```json
{
  "status": "error",
  "managers": [
    {
      "manager_url": "http://localhost:8081",
      "status": "success",
      "http_status": 200
    },
    {
      "manager_url": "https://manager2:8443",
      "status": "error",
      "http_status": 500,
      "error": "unexpected HTTP status: 500"
    }
  ]
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
user = "agent_user"
name = "agent_db"
sslmode = "disable"

[tls]
enabled = false
cert_file = ""
key_file = ""
ca_file = ""

[manager]
urls = ["http://localhost:8081"]
timeout_seconds = 5
```

### Переменные окружения

#### Database
```
DB_HOST=localhost           # PostgreSQL хост
DB_PORT=5432               # PostgreSQL порт
DB_USER=agent_user         # Пользователь БД
DB_PASSWORD=agent_password # Пароль БД
DB_NAME=agent_db           # Имя БД
DB_SSLMODE=disable         # SSL режим (disable/require)
```

#### Application
```
APP_PORT=8080              # Порт HTTP сервера
```

#### Manager (несколько manager-ов через запятую)
```
MANAGER_URLS=https://185.211.170.173:8443,https://92.63.177.186:8443
MANAGER_TIMEOUT=5          # Таймаут в секундах для manager запросов
```

#### TLS (по умолчанию отключен)
```
TLS_ENABLED=false          # Включить TLS (true/false)
TLS_CERT_FILE=             # Путь к сертификату сервера
TLS_KEY_FILE=              # Путь к приватному ключу сервера
TLS_CA_FILE=               # Путь к сертификату CA для проверки client cert
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
- **Масштабируемость**: Поддержка нескольких manager сервисов одновременно

## TLS конфигурация (отключено по умолчанию)

Сервис готов к использованию mTLS (mutual TLS) аутентификации. В данный момент TLS отключен, но архитектура полностью поддерживает включение.

### Включение TLS

Для включения HTTPS и проверки сертификатов клиента:

1. **Подготовить сертификаты:**
   ```bash
   # Сертификат и ключ сервера
   openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes

   # CA сертификат для проверки client cert (опционально для mTLS)
   ```

2. **Установить переменные окружения:**
   ```bash
   export TLS_ENABLED=true
   export TLS_CERT_FILE=/path/to/server.crt
   export TLS_KEY_FILE=/path/to/server.key
   export TLS_CA_FILE=/path/to/ca.crt  # для проверки client cert
   ```

3. **Запустить сервис:**
   ```bash
   go run ./cmd/agent/main.go
   ```

### Параметры TLS

- **TLS_ENABLED**: Включить/выключить TLS (true/false, по умолчанию false)
- **TLS_CERT_FILE**: Путь к файлу с сертификатом сервера (*.crt)
- **TLS_KEY_FILE**: Путь к файлу с приватным ключом сервера (*.key)
- **TLS_CA_FILE**: Путь к файлу с CA сертификатом для проверки client cert (опционально)

## Поддержка нескольких Manager сервисов

Agent может проверять здоровье нескольких manager сервисов одновременно:

### Конфигурация

Укажите несколько URL через запятую в переменной окружения:

```bash
export MANAGER_URLS=https://185.211.170.173:8443,https://92.63.177.186:8443
go run ./cmd/agent/main.go
```

Или в `app.toml`:

```toml
[manager]
urls = [
  "https://185.211.170.173:8443",
  "https://92.63.177.186:8443"
]
timeout_seconds = 5
```

### Результаты проверки

Endpoint `/check-manager` будет проверять все URL и вернёт результаты для каждого:

```json
{
  "status": "error",
  "managers": [
    {
      "manager_url": "https://185.211.170.173:8443",
      "status": "success",
      "http_status": 200
    },
    {
      "manager_url": "https://92.63.177.186:8443",
      "status": "error",
      "http_status": 503,
      "error": "unexpected HTTP status: 503"
    }
  ]
}
```

**Общий статус** (`status`) будет `"error"` если хотя бы один manager недоступен.

## Лицензия

MIT
