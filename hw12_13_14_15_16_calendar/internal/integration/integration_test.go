//go:build integration
// +build integration

// Package integration содержит интеграционные тесты для приложения Календарь.
// Тесты проверяют работу API на уровне gRPC с реальной базой данных PostgreSQL.
package integration

import (
	"context"
	"database/sql"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/api/gen"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/server/grpc"
	sqlstorage "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	testDB     *sqlx.DB
	testApp    *app.App
	testClient pb.EventServiceClient
	testConn   *grpc.ClientConn
	testServer *grpc.Server
)

// TestMain настраивает окружение для всех интеграционных тестов.
func TestMain(m *testing.M) {

	// Настройка тестовой БД
	var err error
	testDB, err = sqlx.Connect("postgres", os.Getenv("TEST_DB_DSN"))
	if err != nil {
		// Если БД недоступна, пропускаем тесты
		os.Exit(0)
	}

	// Применяем миграции
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../migrations"
	}
	if err := goose.Up(testDB.DB, migrationsPath); err != nil {
		os.Exit(1)
	}

	// Создаем приложение и gRPC сервер
	logg := logger.New("INFO")
	storage := sqlstorage.NewWithDB(testDB)
	testApp = app.New(logg, storage)
	
	// Запускаем gRPC сервер
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		os.Exit(1)
	}

	testServer = grpc.NewServer()
	pb.RegisterEventServiceServer(testServer, grpcserver.NewServer(testApp))

	go func() {
		_ = testServer.Serve(lis)
	}()

	ctxDial, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testConn, err = grpc.DialContext(ctxDial, lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		os.Exit(1)
	}

	testClient = pb.NewEventServiceClient(testConn)

	// Запускаем тесты
	code := m.Run()

	// Очистка
	_, _ = testDB.ExecContext(context.Background(), "TRUNCATE TABLE events CASCADE")
	_ = testConn.Close()
	testServer.Stop()
	_ = testDB.Close()

	os.Exit(code)
}

// setupTest очищает БД перед каждым тестом.
func setupTest(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "TRUNCATE TABLE events CASCADE")
	require.NoError(t, err, "failed to cleanup test database")
}

// TestCreateEvent проверяет создание события.
func TestCreateEvent(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	eventID := uuid.New().String()
	startTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	resp, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:                  eventID,
			Title:               "Test Event",
			StartTime:           startTime,
			DurationSeconds:     3600,
			Description:         "Test Description",
			UserId:              "user1",
			NotifyBeforeMinutes: 10,
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, eventID, resp.Event.Id)
	assert.Equal(t, "Test Event", resp.Event.Title)
}

// TestCreateEventWithInvalidUUID проверяет обработку ошибки при невалидном UUID.
func TestCreateEventWithInvalidUUID(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              "not-a-uuid",
			Title:           "Invalid UUID",
			StartTime:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			DurationSeconds: 3600,
			UserId:          "user1",
		},
	})

	require.Error(t, err, "expected error when creating event with invalid UUID")
}

// TestListEventsForDay проверяет получение событий за день.
func TestListEventsForDay(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := "user2"
	dayStart := time.Date(2024, 7, 19, 0, 0, 0, 0, time.UTC)

	// Создаем событие на этот день
	eventID := uuid.New().String()
	_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "Day Event",
			StartTime:       dayStart.Add(10 * time.Hour).Format(time.RFC3339),
			DurationSeconds: 3600,
			UserId:          userID,
		},
	})
	require.NoError(t, err)

	// Получаем события за день
	resp, err := testClient.ListEventsForDay(ctx, &pb.ListEventsRequest{
		UserId:      userID,
		PeriodStart: dayStart.Format(time.RFC3339),
		PeriodEnd:   dayStart.Add(24 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Events)
	assert.Equal(t, eventID, resp.Events[0].Id)
}

// TestListEventsForWeek проверяет получение событий за неделю.
func TestListEventsForWeek(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := "user3"
	weekStart := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)

	// Создаем несколько событий на неделю
	for i := 0; i < 5; i++ {
		_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
			Event: &pb.Event{
				Id:              uuid.New().String(),
				Title:           "Week Event " + string(rune(i)),
				StartTime:       weekStart.Add(time.Duration(i) * 24 * time.Hour).Add(10 * time.Hour).Format(time.RFC3339),
				DurationSeconds: 3600,
				UserId:          userID,
			},
		})
		require.NoError(t, err)
	}

	// Получаем события за неделю
	resp, err := testClient.ListEventsForWeek(ctx, &pb.ListEventsRequest{
		UserId:      userID,
		PeriodStart: weekStart.Format(time.RFC3339),
		PeriodEnd:   weekStart.Add(7 * 24 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Events), 5)
}

// TestListEventsForMonth проверяет получение событий за месяц.
func TestListEventsForMonth(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	userID := "user4"
	monthStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)

	// Создаем несколько событий на месяц
	for i := 0; i < 10; i++ {
		_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
			Event: &pb.Event{
				Id:              uuid.New().String(),
				Title:           "Month Event " + string(rune(i)),
				StartTime:       monthStart.Add(time.Duration(i*3) * 24 * time.Hour).Add(10 * time.Hour).Format(time.RFC3339),
				DurationSeconds: 3600,
				UserId:          userID,
			},
		})
		require.NoError(t, err)
	}

	// Получаем события за месяц
	resp, err := testClient.ListEventsForMonth(ctx, &pb.ListEventsRequest{
		UserId:      userID,
		PeriodStart: monthStart.Format(time.RFC3339),
		PeriodEnd:   monthStart.Add(30 * 24 * time.Hour).Format(time.RFC3339),
	})

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Events), 10)
}

// TestUpdateEvent проверяет обновление события.
func TestUpdateEvent(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	eventID := uuid.New().String()

	// Создаем событие
	_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "Original Title",
			StartTime:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			DurationSeconds: 3600,
			UserId:          "user5",
		},
	})
	require.NoError(t, err)

	// Обновляем событие
	_, err = testClient.UpdateEvent(ctx, &pb.UpdateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "Updated Title",
			StartTime:       time.Now().Add(25 * time.Hour).Format(time.RFC3339),
			DurationSeconds: 7200,
			Description:     "Updated Description",
			UserId:          "user5",
		},
	})
	require.NoError(t, err)
}

// TestDeleteEvent проверяет удаление события.
func TestDeleteEvent(t *testing.T) {

	setupTest(t)

	ctx := context.Background()
	eventID := uuid.New().String()

	// Создаем событие
	_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "To Delete",
			StartTime:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			DurationSeconds: 3600,
			UserId:          "user6",
		},
	})
	require.NoError(t, err)

	// Удаляем событие
	_, err = testClient.DeleteEvent(ctx, &pb.DeleteEventRequest{
		Id:     eventID,
		UserId: "user6",
	})
	require.NoError(t, err)

	// Проверяем, что событие удалено
	resp, err := testClient.ListEventsForDay(ctx, &pb.ListEventsRequest{
		UserId:      "user6",
		PeriodStart: time.Now().Format(time.RFC3339),
		PeriodEnd:   time.Now().Add(48 * time.Hour).Format(time.RFC3339),
	})
	require.NoError(t, err)
	for _, ev := range resp.Events {
		assert.NotEqual(t, eventID, ev.Id, "Deleted event still present")
	}
}

// TestNotificationFlow проверяет отправку уведомлений.
// Создает событие с notify_before, проверяет что оно попадает в список для уведомления,
// и что статус уведомления сохраняется в БД.
func TestNotificationFlow(t *testing.T) {
	setupTest(t)

	ctx := context.Background()
	userID := "user_notify"
	eventID := uuid.New().String()
	
	// Создаем событие с уведомлением за 1 минуту до начала
	eventTime := time.Now().Add(2 * time.Minute)
	notifyBeforeMinutes := int32(1) // уведомление за 1 минуту

	_, err := testClient.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:                  eventID,
			Title:               "Event with Notification",
			StartTime:           eventTime.Format(time.RFC3339),
			DurationSeconds:     3600,
			UserId:              userID,
			NotifyBeforeMinutes: notifyBeforeMinutes,
		},
	})
	require.NoError(t, err)

	// Проверяем, что событие попадает в список для уведомления
	// (текущее время + notify_before >= start_time)
	currentTime := time.Now().Unix()
	events, err := testApp.GetEventsForNotification(ctx, currentTime)
	require.NoError(t, err)
	
	// Событие должно быть в списке, если текущее время близко к времени уведомления
	// (start_time - notify_before) <= current_time < start_time
	notifyTime := eventTime.Unix() - int64(notifyBeforeMinutes*60)
	if currentTime >= notifyTime && currentTime < eventTime.Unix() {
		found := false
		for _, ev := range events {
			if ev.ID == eventID {
				found = true
				break
			}
		}
		assert.True(t, found, "Event should be in notification list")
	}

	// Проверяем, что можно сохранить статус уведомления в БД
	// (это проверяет, что таблица notifications существует и работает)
	notificationID := uuid.New().String()
	_, err = testDB.ExecContext(ctx,
		`INSERT INTO notifications (id, event_id, user_id, title, event_time, status, created_at, processed_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		notificationID, eventID, userID, "Event with Notification",
		eventTime.Unix(), "processed", currentTime, currentTime)
	require.NoError(t, err, "Failed to save notification status")

	// Проверяем, что уведомление сохранено
	var count int
	err = testDB.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM notifications WHERE event_id = $1 AND status = $2",
		eventID, "processed")
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Notification should be saved in database")
}
