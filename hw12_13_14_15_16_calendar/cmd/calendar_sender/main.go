// Package main содержит точку входа для процесса рассыльщика календаря.
// Рассыльщик читает уведомления из очереди RabbitMQ и выводит их в STDOUT.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/config"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/queue"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/sender_config.yaml", "Path to configuration file")
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

	// Создание consumer
	consumer, err := queueConn.Consumer(queueName)
	if err != nil {
		panic("failed to create consumer: " + err.Error())
	}
	defer func() { _ = consumer.Close() }()

	logg.Info("sender started, waiting for notifications...")

	// Настройка graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Обработчик уведомлений
	handler := func(notification queue.Notification) error {
		eventTime := time.Unix(notification.EventTime, 0).Format(time.RFC3339)

		// Вывод уведомления в STDOUT (как требуется в задании)
		fmt.Printf("[NOTIFICATION] Event: %s | Title: %s | User: %s | Time: %s\n",
			notification.EventID,
			notification.Title,
			notification.UserID,
			eventTime,
		)

		// Также логируем
		logg.Info(fmt.Sprintf("notification processed: event_id=%s, user_id=%s, title=%s, time=%s",
			notification.EventID,
			notification.UserID,
			notification.Title,
			eventTime,
		))

		return nil
	}

	// Запуск потребления сообщений
	if err := consumer.Consume(ctx, handler); err != nil {
		logg.Error("failed to consume messages: " + err.Error())
		os.Exit(1)
	}

	logg.Info("sender stopped")
}
