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
	CreateEvent(ctx context.Context, event storage.Event) error                               // Создать событие
	UpdateEvent(ctx context.Context, event storage.Event) error                               // Обновить событие
	DeleteEvent(ctx context.Context, id string) error                                         // Удалить событие
	GetEvent(ctx context.Context, id string) (storage.Event, error)                           // Получить событие по ID
	ListEvents(ctx context.Context, userID string) ([]storage.Event, error)                   // Получить все события пользователя
	GetEventsForNotification(ctx context.Context, currentTime int64) ([]storage.Event, error) // Получить события, требующие уведомления
	DeleteOldEvents(ctx context.Context, beforeTime int64) error                              // Удалить старые события
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

// ListEventsForPeriod возвращает события пользователя в заданном диапазоне времени (Unix timestamp).
func (a *App) ListEventsForPeriod(ctx context.Context, userID string, start, end int64) ([]storage.Event, error) {
	events, err := a.ListEvents(ctx, userID)
	if err != nil {
		return nil, err
	}
	var filtered []storage.Event
	for _, ev := range events {
		if ev.StartTime >= start && ev.StartTime < end {
			filtered = append(filtered, ev)
		}
	}
	return filtered, nil
}

// ListEventsForDay возвращает события пользователя за день.
func (a *App) ListEventsForDay(ctx context.Context, userID string, dayStart int64) ([]storage.Event, error) {
	const daySec = 24 * 60 * 60
	return a.ListEventsForPeriod(ctx, userID, dayStart, dayStart+daySec)
}

// ListEventsForWeek возвращает события пользователя за неделю.
func (a *App) ListEventsForWeek(ctx context.Context, userID string, weekStart int64) ([]storage.Event, error) {
	const weekSec = 7 * 24 * 60 * 60
	return a.ListEventsForPeriod(ctx, userID, weekStart, weekStart+weekSec)
}

// ListEventsForMonth возвращает события пользователя за месяц (30 дней).
func (a *App) ListEventsForMonth(ctx context.Context, userID string, monthStart int64) ([]storage.Event, error) {
	const monthSec = 30 * 24 * 60 * 60
	return a.ListEventsForPeriod(ctx, userID, monthStart, monthStart+monthSec)
}

// Logger возвращает логгер приложения.
func (a *App) Logger() Logger {
	return a.logger
}

// GetEventsForNotification возвращает события, требующие уведомления.
func (a *App) GetEventsForNotification(ctx context.Context, currentTime int64) ([]storage.Event, error) {
	return a.storage.GetEventsForNotification(ctx, currentTime)
}

// DeleteOldEvents удаляет старые события.
func (a *App) DeleteOldEvents(ctx context.Context, beforeTime int64) error {
	return a.storage.DeleteOldEvents(ctx, beforeTime)
}
