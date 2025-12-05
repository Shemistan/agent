# Deployment Guide

## GitHub Actions Deployment

Проект настроен на автоматический деплой на сервер через GitHub Actions при пуше в ветку `main`.

### Настройка GitHub Secrets

Перейдите в Settings → Secrets and variables → Actions и добавьте следующие секреты:

#### SSH Configuration

| Secret Name | Description | Example |
|------------|-------------|---------|
| `SSH_HOST` | IP адрес или домен сервера | `185.211.170.173` |
| `SSH_USER` | Пользователь SSH | `root` или `deploy` |
| `SSH_PRIVATE_KEY` | Приватный SSH ключ | Содержимое `~/.ssh/id_rsa` |

#### Service Configuration

| Secret Name | Description | Example | Required |
|------------|-------------|---------|----------|
| `SERVICE_NAME` | Имя сервиса | `agent` | ✅ |
| `SERVICE_PORT` | Внешний порт сервиса | `8081` | ✅ |
| `APP_PORT` | Внутренний порт приложения | `8081` | ✅ |

#### Database Configuration

| Secret Name | Description | Example | Required |
|------------|-------------|---------|----------|
| `DB_HOST` | Хост БД (внутри Docker) | `db` | ✅ |
| `DB_PORT` | Порт БД (внутри Docker) | `5432` | ✅ |
| `DB_USER` | Пользователь БД | `user` | ✅ |
| `DB_PASSWORD` | Пароль БД | `password` | ✅ |
| `DB_NAME` | Имя БД | `pgsql_db_agent` | ✅ |
| `DB_SSLMODE` | SSL режим | `disable` | ✅ |

⚠️ **Важно**: `DB_HOST` должен быть `db` (имя контейнера), а не `localhost`!

#### Manager Configuration

| Secret Name | Description | Example | Required |
|------------|-------------|---------|----------|
| `MANAGER_URLS` | URLs менеджеров (через запятую) | `http://localhost:8081` | ✅ |
| `MANAGER_TIMEOUT` | Таймаут запросов (секунды) | `5` | ✅ |

#### TLS Configuration (опционально)

| Secret Name | Description | Example | Required |
|------------|-------------|---------|----------|
| `TLS_ENABLED` | Включить TLS | `false` | ✅ |
| `TLS_CERT_FILE` | Путь к сертификату | `/path/to/cert.crt` | ❌ |
| `TLS_KEY_FILE` | Путь к ключу | `/path/to/key.key` | ❌ |
| `TLS_CA_FILE` | Путь к CA сертификату | `/path/to/ca.crt` | ❌ |

#### Docker Hub (опционально)

| Secret Name | Description | Required |
|------------|-------------|----------|
| `DOCKERHUB_USERNAME` | Имя пользователя Docker Hub | ❌ |
| `DOCKERHUB_TOKEN` | Токен Docker Hub | ❌ |

## Процесс деплоя

При пуше в ветку `main` автоматически запускается workflow:

1. **Checkout** - клонирование репозитория
2. **Setup SSH** - настройка SSH агента
3. **Sync to server** - синхронизация файлов на сервер через rsync
4. **Create .env** - создание .env файла из секретов
5. **Deploy** - запуск контейнеров в правильном порядке:
   - Остановка и удаление старых контейнеров
   - Сборка новых образов
   - Запуск БД и ожидание готовности
   - Запуск миграций
   - Запуск приложения

## Последовательность запуска контейнеров

Благодаря `depends_on` в [docker-compose.yml](docker-compose.yml):

```
1. db (PostgreSQL)
   ↓ (waits for healthy)
2. migrator (applies migrations)
   ↓ (waits for successful completion)
3. manager (application service)
```

## Подготовка сервера

### 1. Установка Docker и Docker Compose

```bash
# Обновите систему
apt-get update && apt-get upgrade -y

# Установите Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# Установите Docker Compose
apt-get install docker-compose-plugin
```

### 2. Создание директории для проекта

```bash
mkdir -p /opt/agent
chown -R $USER:$USER /opt/agent
```

### 3. Настройка SSH доступа

```bash
# На локальной машине сгенерируйте SSH ключ (если еще нет)
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"

# Скопируйте публичный ключ на сервер
ssh-copy-id user@server_ip

# Или вручную добавьте в ~/.ssh/authorized_keys на сервере
```

### 4. Открытие портов (если используется firewall)

```bash
# Открыть порт приложения (8081)
ufw allow 8081/tcp

# Если нужен доступ к PostgreSQL извне
ufw allow 54322/tcp

# SSH (если еще не открыт)
ufw allow 22/tcp

# Применить правила
ufw enable
```

## Мониторинг деплоя

### Просмотр логов GitHub Actions

1. Перейдите в репозиторий на GitHub
2. Actions → выберите последний workflow
3. Откройте шаг "Deploy containers"

### Просмотр логов на сервере

```bash
# Подключитесь к серверу
ssh user@server_ip

# Перейдите в директорию проекта
cd /opt/agent

# Просмотр логов всех контейнеров
docker compose logs

# Логи конкретного сервиса
docker compose logs db
docker compose logs migrator
docker compose logs manager

# Следить за логами в реальном времени
docker compose logs -f
```

### Проверка статуса контейнеров

```bash
cd /opt/agent
docker compose ps
```

Ожидаемый результат:
```
NAME             STATUS
pgsql_db_agent   Up (healthy)
agent_app        Up
```

### Проверка работы API

```bash
# Health check
curl http://localhost:8081/health

# Manager check
curl http://localhost:8081/check-manager
```

## Откат изменений

Если деплой прошел неудачно:

```bash
# Подключитесь к серверу
ssh user@server_ip
cd /opt/agent

# Просмотрите логи
docker compose logs

# Остановите контейнеры
docker compose down

# При необходимости очистите volumes
docker compose down -v

# Вернитесь к предыдущей версии кода
git fetch
git checkout <previous_commit_hash>

# Запустите заново
docker compose up -d
```

## Troubleshooting

### Ошибка: "dial tcp: lookup db: server misbehaving"

**Причина**: БД не запущена или мигратор запущен без зависимостей.

**Решение**: Убедитесь, что в GitHub Actions не используется флаг `--no-deps` при запуске docker compose.

### Ошибка: "password authentication failed"

**Причина**: Неправильный пароль БД или несоответствие между .env и секретами GitHub.

**Решение**:
1. Проверьте `DB_PASSWORD` в GitHub Secrets
2. Проверьте файл `/opt/agent/.env` на сервере

### Контейнер постоянно перезапускается

```bash
# Проверьте логи
docker compose logs manager --tail=50

# Проверьте переменные окружения
docker compose exec manager env | grep DB_
```

### Порты заняты

```bash
# Проверьте, что использует порт
lsof -i :8081

# Остановите конфликтующий процесс или измените порт в GitHub Secrets
```

## Полезные команды

```bash
# Остановить все контейнеры
docker compose down

# Остановить и удалить volumes
docker compose down -v

# Пересобрать без кеша
docker compose build --no-cache

# Посмотреть использование ресурсов
docker stats

# Удалить неиспользуемые образы
docker image prune -a

# Зайти в контейнер БД
docker compose exec db psql -U user -d pgsql_db_agent
```
