# Календарь (Calendar Service)

Современный микросервис для хранения и управления событиями календаря. Реализован на Go с чистой архитектурой, поддерживает работу с in-memory и PostgreSQL, автоматические миграции через Goose, конфигурацию через YAML и удобный запуск в Docker.

---

## 🚀 Быстрый старт

### Рекомендуемый подход: База данных в Docker + приложение локально

**Идеально для разработки и тестирования**

1. **Запустить только PostgreSQL:**
   ```bash
   make db-up
   ```

2. **Создать тестовую базу данных:**
   ```bash
   make db-create-test
   ```

3. **Запустить приложение локально:**
   ```bash
   make run
   ```

4. **Тестировать API:**
   ```bash
   curl http://localhost:8080/hello
   ```

5. **Остановить базу данных:**
   ```bash
   make db-down
   ```

### Альтернативный подход: Полный запуск в Docker

**Для продакшена или демонстрации**

1. **Соберите и запустите все сервисы:**
   ```bash
   docker-compose -f docker-compose.full.yml up --build -d
   ```

2. **Проверьте статус:**
   ```bash
   docker-compose -f docker-compose.full.yml ps
   ```

3. **Посмотреть логи:**
   ```bash
   docker-compose -f docker-compose.full.yml logs -f calendar-app
   ```

4. **Остановить сервисы:**
   ```bash
   docker-compose -f docker-compose.full.yml down
   ```

---

## 🛠️ Команды для разработки

### Управление базой данных

```bash
# Запуск только PostgreSQL
make db-up

# Остановка PostgreSQL
make db-down

# Остановка PostgreSQL с удалением данных
make db-down-clean

# Создание тестовой базы данных
make db-create-test

# Просмотр логов базы данных
docker-compose logs postgres
```

### Разработка приложения

```bash
# Сборка приложения
make build

# Запуск приложения локально
make run

# Запуск с кастомной конфигурацией
./bin/calendar --config ./configs/config.yaml

# Запуск с кастомным путем к миграциям
./bin/calendar --migrations ./migrations
```

### Тестирование

```bash
# Обычные тесты (без базы данных)
make test

# Тесты с базой данных в Docker
make test-with-db

# Запуск линтера
make lint
```

---

## 🛠️ Конфигурация

### Основной конфиг (configs/config.yaml)
```yaml
logger:
  level: INFO
storage:
  type: sql
server:
  host: 0.0.0.0
  port: 8080
db:
  host: localhost  # для локального запуска
  port: 5432
  user: calendar
  password: calendar
  dbname: calendar
```

### Конфиг для Docker (configs/config.docker.yaml)
```yaml
logger:
  level: INFO
storage:
  type: sql
server:
  host: 0.0.0.0
  port: 8080
db:
  host: postgres  # для Docker
  port: 5432
  user: calendar
  password: calendar
  dbname: calendar
```

---

## 🗄️ Миграции базы данных (Goose)

- Все миграции хранятся в папке `migrations/` и имеют формат:
  ```
  YYYYMMDDHHMMSS_description.sql
  ```
- Для каждой миграции обязательно наличие секций `-- +goose Up` и `-- +goose Down`.
- Миграции применяются **автоматически** при запуске приложения.

### Пример миграции:
```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    notify_before INTEGER
);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);

-- +goose Down
DROP INDEX IF EXISTS idx_events_start_time;
DROP INDEX IF EXISTS idx_events_user_id;
DROP TABLE IF EXISTS events;
```

### Ручное управление миграциями

- Применить все миграции:
  ```bash
  goose -dir ./migrations postgres "host=localhost port=5432 user=calendar password=calendar dbname=calendar sslmode=disable" up
  ```
- Откатить последнюю миграцию:
  ```bash
  goose -dir ./migrations postgres "host=localhost port=5432 user=calendar password=calendar dbname=calendar sslmode=disable" down
  ```
- Проверить статус:
  ```bash
  goose -dir ./migrations postgres "host=localhost port=5432 user=calendar password=calendar dbname=calendar sslmode=disable" status
  ```

---

## 🐘 Работа с базой данных PostgreSQL

### Подключение к базе данных

```bash
# Войти в контейнер с PostgreSQL
docker-compose exec postgres psql -U calendar -d calendar

# Внутри psql:
# - Посмотреть таблицы: \dt
# - Посмотреть содержимое: SELECT * FROM events;
# - Выйти: \q
```

### Создание тестовой базы

```bash
# Автоматически через make
make db-create-test

# Вручную
docker-compose exec postgres psql -U calendar -c "CREATE DATABASE calendar_test;"
```

### Очистка данных

```bash
# Остановить с удалением данных
make db-down-clean

# Или вручную
docker-compose down -v
```

---

## 🌐 API и тестирование

### Текущие эндпоинты
- `GET /hello` — тестовый эндпоинт

### Тестирование с curl

#### Локальный запуск (порт 8080)
```bash
# Базовый тест
curl http://localhost:8080/hello

# С заголовками
curl -H "Content-Type: application/json" http://localhost:8080/hello

# С verbose режимом
curl -v http://localhost:8080/hello

# POST запрос
curl -X POST http://localhost:8080/hello

# Проверка статуса
curl -I http://localhost:8080/hello
```

#### Docker запуск (порт 8080)
```bash
# Изнутри контейнера
docker exec -it calendar-app curl http://localhost:8080/hello

# С хоста
curl http://localhost:8080/hello
```

### Примеры для будущих API эндпоинтов

```bash
# Создать событие
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Встреча",
    "description": "Важная встреча",
    "user_id": "user123",
    "start_time": "2024-07-15T10:00:00Z",
    "end_time": "2024-07-15T11:00:00Z"
  }'

# Получить все события
curl http://localhost:8080/events

# Получить событие по ID
curl http://localhost:8080/events/123e4567-e89b-12d3-a456-426614174000
```

---

## 📂 Структура проекта

```
hw12_13_14_15_16_calendar/
├── cmd/calendar/           # Точка входа приложения
├── internal/
│   ├── app/               # Бизнес-логика
│   ├── config/            # Конфигурация
│   ├── logger/            # Логирование
│   ├── server/http/       # HTTP API
│   └── storage/           # Хранилища (memory, sql)
├── configs/               # Конфигурационные файлы
│   ├── config.yaml        # Основной конфиг (локальный запуск)
│   └── config.docker.yaml # Конфиг для Docker
├── migrations/            # Миграции Goose
├── build/                 # Dockerfile и сборка
├── docker-compose.yml     # Только база данных (разработка)
├── docker-compose.full.yml # Полный запуск (продакшен)
└── Makefile               # Команды для разработки
```

---

## 🔧 Переменные окружения

### Для тестов
```bash
export TEST_DB_DSN="host=localhost port=5432 user=calendar password=calendar dbname=calendar_test sslmode=disable"
```

### Для приложения
```bash
export CONFIG_FILE="/etc/calendar/config.yaml"  # в Docker
export CONFIG_FILE="./configs/config.yaml"      # локально
```

---

## 🚨 Устранение неполадок

### Проблема: Порт уже занят
```bash
# Изменить порт в configs/config.yaml
server:
  port: 8081  # вместо 8080
```

### Проблема: База данных недоступна
```bash
# Проверить статус контейнера
docker-compose ps

# Перезапустить базу
make db-down
make db-up
```

### Проблема: Миграции не применяются
```bash
# Проверить путь к миграциям
./bin/calendar --migrations ./migrations

# Применить миграции вручную
make db-up
goose -dir ./migrations postgres "host=localhost port=5432 user=calendar password=calendar dbname=calendar sslmode=disable" up
```

---

## 📝 Лицензия
MIT
