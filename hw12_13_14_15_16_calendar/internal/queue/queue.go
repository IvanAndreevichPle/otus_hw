// Package queue содержит абстракцию для работы с очередями сообщений.
// Реализация не зависит от конкретного клиента RabbitMQ.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Notification представляет уведомление о событии.
type Notification struct {
	EventID   string `json:"event_id"`   // ID события
	Title     string `json:"title"`      // заголовок события
	EventTime int64  `json:"event_time"` // время события (Unix timestamp)
	UserID    string `json:"user_id"`    // ID пользователя
}

// Publisher интерфейс для публикации сообщений в очередь.
type Publisher interface {
	Publish(ctx context.Context, notification Notification) error
	Close() error
}

// Consumer интерфейс для потребления сообщений из очереди.
type Consumer interface {
	Consume(ctx context.Context, handler func(Notification) error) error
	Close() error
}

// Connection интерфейс для управления соединением с очередью.
type Connection interface {
	DeclareQueue(ctx context.Context, queueName string) error
	Publisher(queueName string) (Publisher, error)
	Consumer(queueName string) (Consumer, error)
	Close() error
}

// RabbitMQConnection реализует Connection для RabbitMQ.
type RabbitMQConnection struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewConnection создает новое соединение с RabbitMQ.
func NewConnection(url string) (Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQConnection{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// DeclareQueue объявляет очередь в RabbitMQ.
func (r *RabbitMQConnection) DeclareQueue(ctx context.Context, queueName string) error {
	_, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return nil
}

// Publisher возвращает Publisher для публикации сообщений.
func (r *RabbitMQConnection) Publisher(queueName string) (Publisher, error) {
	return &RabbitMQPublisher{
		channel:   r.channel,
		queueName: queueName,
	}, nil
}

// Consumer возвращает Consumer для потребления сообщений.
func (r *RabbitMQConnection) Consumer(queueName string) (Consumer, error) {
	return &RabbitMQConsumer{
		channel:   r.channel,
		queueName: queueName,
	}, nil
}

// Close закрывает соединение с RabbitMQ.
func (r *RabbitMQConnection) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// RabbitMQPublisher реализует Publisher для RabbitMQ.
type RabbitMQPublisher struct {
	channel   *amqp.Channel
	queueName string
}

// Publish публикует уведомление в очередь.
func (p *RabbitMQPublisher) Publish(ctx context.Context, notification Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(ctx,
		"",          // exchange (default)
		p.queueName, // routing key (queue name)
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // сообщение сохраняется на диск
			Body:         body,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close закрывает publisher.
func (p *RabbitMQPublisher) Close() error {
	// Канал закрывается через Connection
	return nil
}

// RabbitMQConsumer реализует Consumer для RabbitMQ.
type RabbitMQConsumer struct {
	channel   *amqp.Channel
	queueName string
}

// Consume начинает потребление сообщений из очереди.
func (c *RabbitMQConsumer) Consume(ctx context.Context, handler func(Notification) error) error {
	// Устанавливаем prefetch для балансировки нагрузки
	err := c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer tag
		false,       // auto-ack (false - ручное подтверждение)
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}

				var notification Notification
				if err := json.Unmarshal(d.Body, &notification); err != nil {
					// Логируем ошибку, но не подтверждаем сообщение
					// В реальном приложении можно отправить в dead letter queue
					d.Nack(false, false) // отклонить без повторной постановки
					continue
				}

				// Обрабатываем уведомление
				if err := handler(notification); err != nil {
					// Ошибка обработки - отклоняем сообщение
					d.Nack(false, true) // отклонить с повторной постановкой
					continue
				}

				// Успешная обработка - подтверждаем сообщение
				d.Ack(false)
			}
		}
	}()

	<-ctx.Done()
	return nil
}

// Close закрывает consumer.
func (c *RabbitMQConsumer) Close() error {
	// Канал закрывается через Connection
	return nil
}

// BuildURL строит URL для подключения к RabbitMQ.
func BuildURL(host string, port int, user, password, vhost string) string {
	if vhost == "" {
		vhost = "/"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%d%s", user, password, host, port, vhost)
}

var (
	ErrConnectionFailed = errors.New("connection to queue failed")
	ErrPublishFailed    = errors.New("failed to publish message")
	ErrConsumeFailed    = errors.New("failed to consume message")
)
