// Package memorystorage предоставляет in-memory реализацию хранилища событий.
// Используется для тестирования и разработки. Данные хранятся в памяти и теряются при перезапуске.
package memorystorage

import (
	"context"
	"sync"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage"
)

// Storage представляет in-memory хранилище событий с потокобезопасным доступом
type Storage struct {
	mu     sync.RWMutex             // мьютекс для синхронизации доступа к данным
	events map[string]storage.Event // карта событий, ключ - ID события
}

// New создает новый экземпляр in-memory хранилища
func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

// CreateEvent создает новое событие в хранилище.
// Проверяет, что время не занято другим событием того же пользователя.
func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверка на занятость времени (простая: совпадение времени старта)
	for _, e := range s.events {
		if e.UserID == event.UserID && e.StartTime == event.StartTime {
			return app.ErrDateBusy
		}
	}
	s.events[event.ID] = event
	return nil
}

// UpdateEvent обновляет существующее событие в хранилище.
// Проверяет существование события и занятость времени другими событиями.
func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, что событие существует
	if _, ok := s.events[event.ID]; !ok {
		return context.Canceled // или custom not found error
	}

	// Проверка на занятость времени (кроме текущего события)
	for id, e := range s.events {
		if id != event.ID && e.UserID == event.UserID && e.StartTime == event.StartTime {
			return app.ErrDateBusy
		}
	}
	s.events[event.ID] = event
	return nil
}

// DeleteEvent удаляет событие по ID из хранилища.
// Возвращает ошибку, если событие не найдено.
func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.events[id]; !ok {
		return context.Canceled // или custom not found error
	}
	delete(s.events, id)
	return nil
}

// GetEvent возвращает событие по ID.
// Возвращает ошибку, если событие не найдено.
func (s *Storage) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[id]
	if !ok {
		return storage.Event{}, context.Canceled // или custom not found error
	}
	return event, nil
}

// ListEvents возвращает все события указанного пользователя.
// Использует read lock для оптимизации производительности.
func (s *Storage) ListEvents(ctx context.Context, userID string) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, e := range s.events {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetEventsForNotification возвращает события, требующие уведомления.
func (s *Storage) GetEventsForNotification(ctx context.Context, currentTime int64) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, e := range s.events {
		if e.NotifyBefore != nil {
			// Событие требует уведомления, если:
			// (start_time - notify_before) <= current_time < start_time
			notifyTime := e.StartTime - *e.NotifyBefore
			if notifyTime <= currentTime && currentTime < e.StartTime {
				result = append(result, e)
			}
		}
	}
	return result, nil
}

// DeleteOldEvents удаляет события, произошедшие более указанного времени назад.
func (s *Storage) DeleteOldEvents(ctx context.Context, beforeTime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, e := range s.events {
		if e.StartTime < beforeTime {
			delete(s.events, id)
		}
	}
	return nil
}
