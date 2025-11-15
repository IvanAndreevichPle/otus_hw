// Package main содержит точку входа для процесса планировщика календаря.
// Планировщик периодически сканирует базу данных, выбирает события для уведомления
// и отправляет их в очередь RabbitMQ, а также очищает старые события.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/config"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/queue"
	sqlstorage "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	configFile     string
	migrationsPath string
)

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.yaml", "Path to configuration file")
	flag.StringVar(&migrationsPath, "migrations", "migrations", "Path to migrations directory")
}

func main() {
	flag.Parse()

	// Чтение конфигурации
	cfg, err := config.NewConfigFromFile(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// Инициализация логгера
	logg := logger.New(cfg.Logger.Level)

	// Подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}
	defer func() { _ = db.Close() }()

	// Применение миграций
	if err := runMigrations(db); err != nil {
		panic("failed to apply migrations: " + err.Error())
	}

	// Создание хранилища
	storage := sqlstorage.NewWithDB(sqlx.NewDb(db, "postgres"))

	// Подключение к RabbitMQ
	rabbitURL := queue.BuildURL(
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.VHost,
	)

	queueConn, err := queue.NewConnection(rabbitURL)
	if err != nil {
		panic("failed to connect to RabbitMQ: " + err.Error())
	}
	defer func() { _ = queueConn.Close() }()

	// Объявление очереди
	queueName := cfg.RabbitMQ.Queue
	if queueName == "" {
		queueName = "notifications"
	}

	ctx := context.Background()
	if err := queueConn.DeclareQueue(ctx, queueName); err != nil {
		panic("failed to declare queue: " + err.Error())
	}

	// Создание publisher
	publisher, err := queueConn.Publisher(queueName)
	if err != nil {
		panic("failed to create publisher: " + err.Error())
	}
	defer func() { _ = publisher.Close() }()

	// Инициализация приложения
	calendarApp := app.New(logg, storage)

	// Настройка интервала проверки
	interval := time.Duration(cfg.Scheduler.IntervalSeconds) * time.Second
	if interval == 0 {
		interval = 60 * time.Second // по умолчанию 60 секунд
	}

	logg.Info(fmt.Sprintf("scheduler started with interval %v", interval))

	// Настройка graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Запуск периодической проверки
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первый запуск сразу
	processNotifications(ctx, logg, calendarApp, publisher, queueName)

	// Периодический запуск
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				processNotifications(ctx, logg, calendarApp, publisher, queueName)
			}
		}
	}()

	<-ctx.Done()
	logg.Info("scheduler stopped")
}

// processNotifications обрабатывает уведомления: выбирает события и отправляет в очередь.
func processNotifications(ctx context.Context, logg app.Logger, calendarApp *app.App, publisher queue.Publisher, queueName string) {
	currentTime := time.Now().Unix()

	// Получение событий, требующих уведомления
	events, err := calendarApp.GetEventsForNotification(ctx, currentTime)
	if err != nil {
		logg.Error(fmt.Sprintf("failed to get events for notification: %v", err))
		return
	}

	logg.Info(fmt.Sprintf("found %d events for notification", len(events)))

	// Отправка уведомлений в очередь
	for _, event := range events {
		notification := queue.Notification{
			EventID:   event.ID,
			Title:     event.Title,
			EventTime: event.StartTime,
			UserID:    event.UserID,
		}

		if err := publisher.Publish(ctx, notification); err != nil {
			logg.Error(fmt.Sprintf("failed to publish notification for event %s: %v", event.ID, err))
			continue
		}

		logg.Info(fmt.Sprintf("notification sent for event %s (user: %s, time: %d)", event.ID, event.UserID, event.StartTime))
	}

	// Очистка старых событий (более 1 года назад)
	oneYearAgo := currentTime - 365*24*60*60
	if err := calendarApp.DeleteOldEvents(ctx, oneYearAgo); err != nil {
		logg.Error(fmt.Sprintf("failed to delete old events: %v", err))
	} else {
		logg.Info("old events cleanup completed")
	}
}

// runMigrations применяет миграции базы данных.
func runMigrations(db *sql.DB) error {
	return goose.Up(db, migrationsPath)
}
