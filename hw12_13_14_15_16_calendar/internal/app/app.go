// Package app содержит бизнес-логику приложения Календарь и абстракции для логгера и хранилища.
package app

import (
	"context"
	"errors"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage"
)

// App — основной сервис календаря, объединяющий бизнес-логику, логгер и хранилище.
type App struct {
	logger  Logger  // интерфейс логгера
	storage Storage // интерфейс хранилища событий
}

// Logger — интерфейс для логирования событий приложения.
type Logger interface {
	Info(msg string)  // Информационные сообщения
	Error(msg string) // Ошибки
	Warn(msg string)  // Предупреждения
	Debug(msg string) // Отладочная информация
}

// Storage — интерфейс для работы с хранилищем событий.
type Storage interface {
	CreateEvent(ctx context.Context, event storage.Event) error             // Создать событие
	UpdateEvent(ctx context.Context, event storage.Event) error             // Обновить событие
	DeleteEvent(ctx context.Context, id string) error                       // Удалить событие
	GetEvent(ctx context.Context, id string) (storage.Event, error)         // Получить событие по ID
	ListEvents(ctx context.Context, userID string) ([]storage.Event, error) // Получить все события пользователя
}

// ErrDateBusy — ошибка, если время уже занято другим событием.
var ErrDateBusy = errors.New("date is busy by another event")

// New создает новый экземпляр App.
func New(logger Logger, storage Storage) *App {
	return &App{logger: logger, storage: storage}
}

// CreateEvent создает новое событие в хранилище.
func (a *App) CreateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.CreateEvent(ctx, event)
}

// UpdateEvent обновляет существующее событие.
func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.UpdateEvent(ctx, event)
}

// DeleteEvent удаляет событие по ID.
func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.DeleteEvent(ctx, id)
}

// GetEvent возвращает событие по ID.
func (a *App) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	return a.storage.GetEvent(ctx, id)
}

// ListEvents возвращает все события пользователя.
func (a *App) ListEvents(ctx context.Context, userID string) ([]storage.Event, error) {
	return a.storage.ListEvents(ctx, userID)
}
