// Package main содержит точку входа в приложение календаря.
// Основная функция инициализирует все компоненты системы и запускает HTTP-сервер.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	"github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/config"
	"github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/IvanAndreevichPle/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// configFile содержит путь к файлу конфигурации, заданный через флаг командной строки
var configFile string

// migrationsPath содержит путь к директории с миграциями
var migrationsPath string

// init инициализирует флаги командной строки
func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
	flag.StringVar(&migrationsPath, "migrations", "migrations", "Path to migrations directory")
}

// main является точкой входа в приложение календаря.
// Выполняет инициализацию всех компонентов системы и запускает HTTP-сервер.
func main() {
	flag.Parse()

	// Обработка команды version для отображения версии приложения
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// Чтение и парсинг конфигурации из файла
	configData, err := config.NewConfigFromFile(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Инициализация логгера с уровнем из конфигурации
	logg := logger.New(configData.Logger.Level)

	// Инициализация хранилища в зависимости от типа, указанного в конфигурации
	var storage app.Storage
	switch configData.Storage.Type {
	case "memory":
		// Использование in-memory хранилища для тестирования
		storage = memorystorage.New()
	case "sql":
		// Формируем DSN (Data Source Name) для подключения к PostgreSQL
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			configData.DB.Host, configData.DB.Port, configData.DB.User, configData.DB.Password, configData.DB.DBName)

		// Автоматическое применение миграций при запуске
		if err := runMigrations(dsn); err != nil {
			panic("failed to apply migrations: " + err.Error())
		}

		// Создание SQL хранилища с подключением к базе данных
		storage, err = sqlstorage.New(dsn)
		if err != nil {
			panic("failed to connect to db: " + err.Error())
		}
	default:
		panic("unknown storage type: " + configData.Storage.Type)
	}

	// Инициализация бизнес-логики приложения с логгером и хранилищем
	calendar := app.New(logg, storage)

	// Создание и настройка HTTP-сервера
	server := internalhttp.NewServer(logg, calendar, configData.Server.Host, configData.Server.Port)

	// Настройка graceful shutdown через обработку системных сигналов
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Горутина для graceful shutdown сервера при получении сигнала
	go func() {
		<-ctx.Done()

		// Даем серверу 3 секунды на завершение работы
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	// Запуск HTTP-сервера
	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}

// runMigrations применяет миграции базы данных с помощью Goose.
// Функция подключается к PostgreSQL и применяет все новые миграции.
func runMigrations(dsn string) error {
	// Открытие соединения с базой данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	// Отладочная информация для диагностики проблем с миграциями
	fmt.Printf("DSN: %s\n", dsn)
	fmt.Printf("Migrations path: %s\n", migrationsPath)

	// Проверяем, что файлы миграций существуют в контейнере
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		fmt.Printf("Error reading migrations directory: %v\n", err)
	} else {
		fmt.Printf("Files in migrations directory:\n")
		for _, file := range files {
			fmt.Printf("  - %s\n", file.Name())
		}
	}

	// Применяем миграции с помощью Goose
	if err := goose.Up(db, migrationsPath); err != nil {
		return err
	}
	return nil
}
