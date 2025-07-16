package sqlstorage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func getTestDSN() string {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=calendar password=calendar dbname=calendar_test sslmode=disable"
	}
	return dsn
}

func setupTestStorage(t *testing.T) *Storage {
	dsn := getTestDSN()
	// Автоматически применяем миграции к тестовой базе
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db for migrations: %v", err)
	}
	// Используем правильный путь к миграциям - из корня проекта
	migrationsDir, err := filepath.Abs("../../../migrations")
	if err != nil {
		t.Fatalf("failed to resolve migrations path: %v", err)
	}
	t.Logf("Migrations directory: %s", migrationsDir)
	if err := goose.Up(db, migrationsDir); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	_ = db.Close()

	s, err := New(dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	// Очищаем таблицу перед тестом
	_, _ = s.db.Exec("DELETE FROM events")
	return s
}

func TestSQLStorageCRUD(t *testing.T) {
	s := setupTestStorage(t)
	ctx := context.Background()
	event := storage.Event{
		Title:       "Test Event",
		Description: "desc",
		UserID:      "user1",
		StartTime:   time.Now().Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
	}
	// Create
	if err := s.CreateEvent(ctx, event); err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}
	// List
	list, err := s.ListEvents(ctx, event.UserID)
	if err != nil || len(list) == 0 {
		t.Fatalf("ListEvents failed: %v", err)
	}
	id := list[0].ID
	// Get
	got, err := s.GetEvent(ctx, id)
	if err != nil || got.Title != event.Title {
		t.Fatalf("GetEvent failed: %v", err)
	}
	// Update
	got.Title = "Updated"
	if err := s.UpdateEvent(ctx, got); err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}
	// Delete
	if err := s.DeleteEvent(ctx, id); err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}
	_, err = s.GetEvent(ctx, id)
	if err == nil {
		t.Fatalf("GetEvent should fail after delete")
	}
}

func TestSQLStorageNotFound(t *testing.T) {
	s := setupTestStorage(t)
	ctx := context.Background()
	_, err := s.GetEvent(ctx, "nonexistent")
	if err == nil {
		t.Fatalf("expected not found error")
	}
}
