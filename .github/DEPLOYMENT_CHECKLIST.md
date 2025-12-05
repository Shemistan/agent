# Deployment Checklist

## Перед первым деплоем

### ✅ GitHub Secrets - Обязательные

Убедитесь, что все секреты добавлены в Settings → Secrets and variables → Actions:

- [ ] `SSH_HOST` - IP/домен сервера
- [ ] `SSH_USER` - SSH пользователь
- [ ] `SSH_PRIVATE_KEY` - Приватный SSH ключ

**Service:**
- [ ] `SERVICE_NAME` = `agent`
- [ ] `SERVICE_PORT` = `8081`
- [ ] `APP_PORT` = `8081`

**Database:**
- [ ] `DB_HOST` = `db` ⚠️ **НЕ localhost!**
- [ ] `DB_PORT` = `5432` ⚠️ **Внутренний порт контейнера!**
- [ ] `DB_USER` = `user`
- [ ] `DB_PASSWORD` = ваш_пароль
- [ ] `DB_NAME` = `pgsql_db_agent`
- [ ] `DB_SSLMODE` = `disable`

**Manager:**
- [ ] `MANAGER_URLS` = `http://localhost:8081`
- [ ] `MANAGER_TIMEOUT` = `5`

**TLS:**
- [ ] `TLS_ENABLED` = `false`

### ✅ Подготовка сервера

- [ ] Docker установлен (`docker --version`)
- [ ] Docker Compose установлен (`docker compose version`)
- [ ] Создана директория `/opt/agent`
- [ ] SSH доступ работает (`ssh user@server_ip`)
- [ ] Порт 8081 открыт (если используется firewall)

### ✅ Проверка конфигурации

**Локально:**
```bash
# Проверьте docker-compose.yml
docker compose config | grep -A 3 "DB_HOST"

# Должен быть DB_HOST: db для migrator и manager
```

**docker-compose.yml:**
- [ ] `DB_HOST: db` для migrator (строка 26)
- [ ] `DB_HOST: db` для manager (строка 46)
- [ ] `DB_PORT: 5432` для migrator (строка 27)
- [ ] `DB_PORT: 5432` для manager (строка 47)
- [ ] `depends_on` правильно настроен

## После деплоя

### ✅ Проверка статуса

На сервере:
```bash
ssh user@server_ip
cd /opt/agent

# 1. Проверить контейнеры
docker compose ps

# Должно быть:
# pgsql_db_agent - Up (healthy)
# agent_app - Up

# 2. Проверить логи
docker compose logs --tail=50

# Должны быть:
# - "Starting migrator"
# - "Connected to database"
# - "Migrations completed successfully"
# - "Starting agent service"
# - "Starting HTTP server"
```

### ✅ Проверка работы API

```bash
# Health check
curl http://localhost:8081/health
# Ожидается: {"status":"success"}

# Manager check
curl http://localhost:8081/check-manager
# Ожидается: {"status":"success","managers":[...]}
```

### ✅ Проверка БД

```bash
docker compose exec db psql -U user -d pgsql_db_agent -c "SELECT COUNT(*) FROM health_calls;"
docker compose exec db psql -U user -d pgsql_db_agent -c "SELECT COUNT(*) FROM manager_checks;"
```

## Типичные ошибки

### ❌ "dial tcp: lookup db: server misbehaving"

**Причина**: БД не запущена или мигратор запущен без зависимостей.

**Решение**:
- Убедитесь, что в deploy.yml используется `docker compose up -d` без `--no-deps`
- Проверьте `depends_on` в docker-compose.yml

### ❌ "password authentication failed"

**Причина**: Неправильный пароль или несоответствие секретов.

**Решение**:
- Проверьте `DB_PASSWORD` в GitHub Secrets
- Проверьте `/opt/agent/.env` на сервере

### ❌ "connection refused"

**Причина**: Контейнер не запущен или порт не открыт.

**Решение**:
- `docker compose ps` - проверьте статус
- `docker compose logs manager` - проверьте логи
- Откройте порт в firewall: `ufw allow 8081/tcp`

## Мониторинг

### Логи в реальном времени

```bash
ssh user@server_ip
cd /opt/agent
docker compose logs -f
```

### Использование ресурсов

```bash
docker stats
```

### Проверка через GitHub Actions

1. Откройте Actions в репозитории
2. Выберите последний workflow
3. Проверьте шаги "Deploy containers"
4. В конце должны быть логи контейнеров без ошибок

## Откат

Если что-то пошло не так:

```bash
ssh user@server_ip
cd /opt/agent

# Остановить контейнеры
docker compose down

# Вернуться к предыдущей версии
git checkout <previous_commit>

# Запустить заново
docker compose up -d
```

---

**Полная документация**: [DEPLOYMENT.md](../DEPLOYMENT.md)
