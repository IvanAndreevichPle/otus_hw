# Dockerfile для сборки приложения Календарь
# Использует multi-stage build для создания минимального образа

# Этап сборки: компиляция Go приложения
FROM golang:1.23 as build

# Переменные окружения для сборки
ENV BIN_FILE /opt/calendar/calendar-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Копируем файлы зависимостей и кэшируем слои с модулями
# Это оптимизирует сборку при изменении только исходного кода
COPY go.mod .
COPY go.sum .
RUN go mod download

# Копируем весь исходный код
COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
# LDFLAGS передается из Makefile для внедрения информации о версии
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/calendar/*

# Этап финального образа: минимальный Alpine Linux
FROM alpine:3.9

# Метаданные образа
LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

# Копируем скомпилированный бинарник из этапа сборки
ENV BIN_FILE "/opt/calendar/calendar-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

# Копируем конфигурацию и миграции
ENV CONFIG_FILE /etc/calendar/config.yaml
COPY ./configs/config.yaml ${CONFIG_FILE}
COPY ./migrations /migrations
WORKDIR /migrations

# Команда запуска приложения
CMD ${BIN_FILE} -config ${CONFIG_FILE}
