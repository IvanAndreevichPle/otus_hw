// Package sqlstorage предоставляет PostgreSQL реализацию хранилища событий.
// Использует sqlx для удобной работы с SQL запросами и PostgreSQL.
package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Ошибки, специфичные для SQL хранилища
var (
	ErrNotFound   = errors.New("event not found")  // событие не найдено
	ErrValidation = errors.New("validation error") // ошибка валидации
)

// Storage представляет PostgreSQL хранилище событий
type Storage struct {
	db *sqlx.DB // подключение к базе данных
}

// New создает новое подключение к PostgreSQL и возвращает SQL хранилище
func New(dsn string) (*Storage, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

// Close закрывает соединение с базой данных
func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close()
}

// CreateEvent создает новое событие в базе данных.
// Автоматически генерирует UUID, если ID не указан.
func (s *Storage) CreateEvent(ctx context.Context, event storage.Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO events (id, title, description, user_id, start_time, end_time, notify_before) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		event.ID, event.Title, event.Description, event.UserID, toTime(event.StartTime), toTime(event.EndTime), event.NotifyBefore)
	return err
}

// UpdateEvent обновляет существующее событие в базе данных.
// Возвращает ErrNotFound, если событие не найдено.
func (s *Storage) UpdateEvent(ctx context.Context, event storage.Event) error {
	res, err := s.db.ExecContext(ctx, `UPDATE events SET title=$1, description=$2, user_id=$3, start_time=$4, end_time=$5, notify_before=$6 WHERE id=$7`,
		event.Title, event.Description, event.UserID, toTime(event.StartTime), toTime(event.EndTime), event.NotifyBefore, event.ID)
	if err != nil {
		return err
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteEvent удаляет событие по ID из базы данных.
// Возвращает ErrNotFound, если событие не найдено.
func (s *Storage) DeleteEvent(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id=$1`, id)
	if err != nil {
		return err
	}
	cnt, _ := res.RowsAffected()
	if cnt == 0 {
		return ErrNotFound
	}
	return nil
}

// GetEvent возвращает событие по ID из базы данных.
// Возвращает ErrNotFound, если событие не найдено.
// Обрабатывает nullable поле notify_before.
func (s *Storage) GetEvent(ctx context.Context, id string) (storage.Event, error) {
	var e storage.Event
	row := s.db.QueryRowxContext(ctx, `SELECT id, title, description, user_id, EXTRACT(EPOCH FROM to_timestamp(start_time))::bigint, EXTRACT(EPOCH FROM to_timestamp(end_time))::bigint, notify_before FROM events WHERE id=$1`, id)
	var notifyBefore sql.NullInt64
	if err := row.Scan(&e.ID, &e.Title, &e.Description, &e.UserID, &e.StartTime, &e.EndTime, &notifyBefore); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e, ErrNotFound
		}
		return e, err
	}
	if notifyBefore.Valid {
		e.NotifyBefore = &notifyBefore.Int64
	}
	return e, nil
}

// ListEvents возвращает все события указанного пользователя.
// Обрабатывает nullable поле notify_before для каждого события.
func (s *Storage) ListEvents(ctx context.Context, userID string) ([]storage.Event, error) {
	var events []storage.Event
	rows, err := s.db.QueryxContext(ctx, `SELECT id, title, description, user_id, EXTRACT(EPOCH FROM to_timestamp(start_time))::bigint, EXTRACT(EPOCH FROM to_timestamp(end_time))::bigint, notify_before FROM events WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var e storage.Event
		var notifyBefore sql.NullInt64
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.UserID, &e.StartTime, &e.EndTime, &notifyBefore); err != nil {
			return nil, err
		}
		if notifyBefore.Valid {
			e.NotifyBefore = &notifyBefore.Int64
		}
		events = append(events, e)
	}
	return events, nil
}

// toTime конвертирует Unix timestamp в строку для PostgreSQL
func toTime(ts int64) string {
	return fmt.Sprintf("%d", ts)
}
