# Makefile для проекта Календарь
# Содержит команды для сборки, тестирования и развертывания приложения

# Переменные для сборки
BIN := "./bin/calendar"                    # путь к исполняемому файлу
DOCKER_IMG="calendar:develop"              # имя Docker образа

# Информация о сборке для внедрения в бинарный файл
GIT_HASH := $(shell git log --format="%h" -n 1)  # хеш последнего коммита
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

# Сборка исполняемого файла с внедренной информацией о версии
build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

# Сборка и запуск приложения с конфигурацией по умолчанию
run: build
	$(BIN) --config ./configs/config.yaml

# Сборка Docker образа
build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

# Сборка и запуск Docker образа
run-img: build-img
	docker run $(DOCKER_IMG)

# Запуск только базы данных PostgreSQL в Docker
db-up:
	docker-compose up -d

# Остановка базы данных PostgreSQL
db-down:
	docker-compose down

# Остановка базы данных PostgreSQL с удалением данных
db-down-clean:
	docker-compose down -v

# Создание тестовой базы данных
db-create-test:
	docker-compose exec postgres psql -U calendar -c "CREATE DATABASE calendar_test;"

# Запуск тестов с базой данных в Docker
test-with-db: db-up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@docker-compose exec postgres psql -U calendar -c "CREATE DATABASE IF NOT EXISTS calendar_test;" 2>/dev/null || true
	@make test
	@make db-down

# Отображение версии приложения
version: build
	$(BIN) version

# Запуск тестов с проверкой на race conditions
test:
	go test -race ./internal/...

# Установка зависимостей для линтера (если не установлены)
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.63.4

mod-download:
	go mod download

# Запуск линтера для проверки качества кода
lint: mod-download install-lint-deps
	golangci-lint run ./...

# Генерация Go-кода из Protobuf-спецификаций

generate:
	cd api && go generate gen.go

# Объявление phony targets (целей, которые не являются файлами)
.PHONY: build run build-img run-img version test lint db-up db-down db-down-clean db-create-test test-with-db generate