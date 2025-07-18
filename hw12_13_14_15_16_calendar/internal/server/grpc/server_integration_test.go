package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/api/gen"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/server/grpc"
	memorystorage "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func startTestGRPCServer(t *testing.T) (pb.EventServiceClient, func()) {
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	s := grpc.NewServer()
	appInstance := app.New(logger.New("DEBUG"), memorystorage.New())
	pb.RegisterEventServiceServer(s, grpcserver.NewServer(appInstance))

	go func() {
		_ = s.Serve(lis)
	}()

	ctxDial, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//nolint:staticcheck // grpc.DialContext is deprecated, but required for current gRPC version
	conn, err := grpc.DialContext(ctxDial, lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)

	client := pb.NewEventServiceClient(conn)
	return client, func() {
		if err := conn.Close(); err != nil {
			t.Logf("conn.Close error: %v", err)
		}
		if err := lis.Close(); err != nil {
			t.Logf("lis.Close error: %v", err)
		}
		s.Stop()
	}
}

func TestCreateAndListEvent(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	eventID := uuid.NewString()
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:                  eventID,
			Title:               "Integration Test",
			StartTime:           "2024-07-19T10:00:00Z",
			DurationSeconds:     3600,
			Description:         "Интеграционный тест",
			UserId:              "user1",
			NotifyBeforeMinutes: 5,
		},
	})
	require.NoError(t, err)

	resp, err := client.ListEventsForDay(ctx, &pb.ListEventsRequest{
		UserId:      "user1",
		PeriodStart: "2024-07-19T00:00:00Z",
		PeriodEnd:   "2024-07-20T00:00:00Z",
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Events)
}

func TestUpdateEvent(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	eventID := uuid.NewString()
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "ToUpdate",
			StartTime:       "2024-07-20T10:00:00Z",
			DurationSeconds: 1800,
			Description:     "Before update",
			UserId:          "user2",
		},
	})
	require.NoError(t, err)

	_, err = client.UpdateEvent(ctx, &pb.UpdateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "Updated",
			StartTime:       "2024-07-20T11:00:00Z",
			DurationSeconds: 3600,
			Description:     "After update",
			UserId:          "user2",
		},
	})
	require.NoError(t, err)

	resp, err := client.ListEventsForDay(ctx, &pb.ListEventsRequest{
		UserId:      "user2",
		PeriodStart: "2024-07-20T00:00:00Z",
		PeriodEnd:   "2024-07-21T00:00:00Z",
	})
	require.NoError(t, err)
	found := false
	for _, ev := range resp.Events {
		if ev.Id == eventID && ev.Title == "Updated" {
			found = true
		}
	}
	require.True(t, found, "Updated event not found")
}

func TestDeleteEvent(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	eventID := uuid.NewString()
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              eventID,
			Title:           "ToDelete",
			StartTime:       "2024-07-21T10:00:00Z",
			DurationSeconds: 1200,
			UserId:          "user3",
		},
	})
	require.NoError(t, err)

	_, err = client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: eventID, UserId: "user3"})
	require.NoError(t, err)

	resp, err := client.ListEventsForDay(ctx, &pb.ListEventsRequest{
		UserId:      "user3",
		PeriodStart: "2024-07-21T00:00:00Z",
		PeriodEnd:   "2024-07-22T00:00:00Z",
	})
	require.NoError(t, err)
	for _, ev := range resp.Events {
		require.NotEqual(t, eventID, ev.Id, "Deleted event still present")
	}
}

func TestListEventsForWeekAndMonth(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	userID := "user4"
	// Создаём события на разные дни
	for i := 0; i < 10; i++ {
		start := time.Date(2024, 7, 10+i, 10, 0, 0, 0, time.UTC)
		_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
			Event: &pb.Event{
				Id:              uuid.NewString(),
				Title:           "Event " + time.Now().String(),
				StartTime:       start.Format(time.RFC3339),
				DurationSeconds: 3600,
				UserId:          userID,
			},
		})
		require.NoError(t, err)
	}

	// Неделя с 10 по 17 июля
	respWeek, err := client.ListEventsForWeek(ctx, &pb.ListEventsRequest{
		UserId:      userID,
		PeriodStart: "2024-07-10T00:00:00Z",
		PeriodEnd:   "2024-07-17T00:00:00Z",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(respWeek.Events), 7)

	// Месяц с 1 по 31 июля
	respMonth, err := client.ListEventsForMonth(ctx, &pb.ListEventsRequest{
		UserId:      userID,
		PeriodStart: "2024-07-01T00:00:00Z",
		PeriodEnd:   "2024-08-01T00:00:00Z",
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(respMonth.Events), 10)
}

func TestDeleteNonExistentEvent(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	nonExistentID := uuid.NewString()
	_, err := client.DeleteEvent(ctx, &pb.DeleteEventRequest{Id: nonExistentID, UserId: "userX"})
	require.Error(t, err, "expected error when deleting non-existent event")
}

func TestUpdateNonExistentEvent(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	nonExistentID := uuid.NewString()
	_, err := client.UpdateEvent(ctx, &pb.UpdateEventRequest{
		Event: &pb.Event{
			Id:              nonExistentID,
			Title:           "ShouldNotExist",
			StartTime:       "2024-07-25T10:00:00Z",
			DurationSeconds: 3600,
			UserId:          "userY",
		},
	})
	require.Error(t, err, "expected error when updating non-existent event")
}

func TestCreateEventWithInvalidUUID(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              "not-a-uuid",
			Title:           "Invalid UUID",
			StartTime:       "2024-07-25T10:00:00Z",
			DurationSeconds: 3600,
			UserId:          "userZ",
		},
	})
	require.Error(t, err, "expected error when creating event with invalid UUID")
}

// Если бизнес-логика запрещает пересечение событий, этот тест будет актуален.
func TestCreateOverlappingEvents(t *testing.T) {
	client, cleanup := startTestGRPCServer(t)
	defer cleanup()

	ctx := context.Background()
	userID := "userOverlap"
	startTime := "2024-07-26T10:00:00Z"

	// Первое событие
	_, err := client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              uuid.NewString(),
			Title:           "First",
			StartTime:       startTime,
			DurationSeconds: 3600,
			UserId:          userID,
		},
	})
	require.NoError(t, err)

	// Пересекающееся событие
	_, err = client.CreateEvent(ctx, &pb.CreateEventRequest{
		Event: &pb.Event{
			Id:              uuid.NewString(),
			Title:           "Overlap",
			StartTime:       startTime,
			DurationSeconds: 1800,
			UserId:          userID,
		},
	})
	// Если пересечения запрещены, ожидаем ошибку. Если разрешены — замените на require.NoError.
	require.Error(t, err, "expected error on overlapping event")
}
