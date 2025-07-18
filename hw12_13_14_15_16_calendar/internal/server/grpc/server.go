package grpc

import (
	context "context"
	"time"

	pb "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/api/gen"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage"
)

// Server реализует pb.EventServiceServer и связывает GRPC с бизнес-логикой.
type Server struct {
	pb.UnimplementedEventServiceServer
	app *app.App
}

// NewServer создаёт новый grpc-сервер с внедрённой бизнес-логикой.
func NewServer(app *app.App) *Server {
	return &Server{app: app}
}

// CreateEvent реализует метод создания события через GRPC.
func (s *Server) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	event := req.GetEvent()
	// Логирование
	s.app.Logger().Info("GRPC CreateEvent: " + event.GetTitle())

	// Маппинг pb.Event -> storage.Event
	storageEvent, err := protoToStorageEvent(event)
	if err != nil {
		s.app.Logger().Error("CreateEvent mapping error: " + err.Error())
		return nil, err
	}

	err = s.app.CreateEvent(ctx, storageEvent)
	if err != nil {
		s.app.Logger().Error("CreateEvent error: " + err.Error())
		return nil, err
	}

	return &pb.CreateEventResponse{Event: event}, nil
}

// UpdateEvent реализует обновление события через GRPC.
func (s *Server) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	event := req.GetEvent()
	s.app.Logger().Info("GRPC UpdateEvent: " + event.GetTitle())
	storageEvent, err := protoToStorageEvent(event)
	if err != nil {
		s.app.Logger().Error("UpdateEvent mapping error: " + err.Error())
		return nil, err
	}
	err = s.app.UpdateEvent(ctx, storageEvent)
	if err != nil {
		s.app.Logger().Error("UpdateEvent error: " + err.Error())
		return nil, err
	}
	return &pb.UpdateEventResponse{Event: event}, nil
}

// DeleteEvent реализует удаление события через GRPC.
func (s *Server) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	s.app.Logger().Info("GRPC DeleteEvent: " + req.GetId())
	err := s.app.DeleteEvent(ctx, req.GetId())
	if err != nil {
		s.app.Logger().Error("DeleteEvent error: " + err.Error())
		return nil, err
	}
	return &pb.DeleteEventResponse{Success: true}, nil
}

// ListEventsForDay реализует получение событий за день через GRPC.
func (s *Server) ListEventsForDay(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	return s.listEventsForPeriod(ctx, req, "day")
}

// ListEventsForWeek реализует получение событий за неделю через GRPC.
func (s *Server) ListEventsForWeek(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	return s.listEventsForPeriod(ctx, req, "week")
}

// ListEventsForMonth реализует получение событий за месяц через GRPC.
func (s *Server) ListEventsForMonth(ctx context.Context, req *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	return s.listEventsForPeriod(ctx, req, "month")
}

// listEventsForPeriod — вспомогательный метод для выборки событий по периоду.
func (s *Server) listEventsForPeriod(ctx context.Context, req *pb.ListEventsRequest, period string) (*pb.ListEventsResponse, error) {
	s.app.Logger().Info("GRPC ListEventsFor" + period + ": " + req.GetUserId())
	start, err := time.Parse(time.RFC3339, req.GetPeriodStart())
	if err != nil {
		s.app.Logger().Error("ListEventsFor" + period + " start parse error: " + err.Error())
		return nil, err
	}
	var events []storage.Event
	switch period {
	case "day":
		events, err = s.app.ListEventsForDay(ctx, req.GetUserId(), start.Unix())
	case "week":
		events, err = s.app.ListEventsForWeek(ctx, req.GetUserId(), start.Unix())
	case "month":
		events, err = s.app.ListEventsForMonth(ctx, req.GetUserId(), start.Unix())
	}
	if err != nil {
		s.app.Logger().Error("ListEventsFor" + period + " error: " + err.Error())
		return nil, err
	}
	// Маппинг storage.Event -> pb.Event
	var pbEvents []*pb.Event
	for _, ev := range events {
		pbEvents = append(pbEvents, storageToProtoEvent(ev))
	}
	return &pb.ListEventsResponse{Events: pbEvents}, nil
}

// protoToStorageEvent преобразует pb.Event в storage.Event
func protoToStorageEvent(e *pb.Event) (storage.Event, error) {
	start, err := time.Parse(time.RFC3339, e.GetStartTime())
	if err != nil {
		return storage.Event{}, err
	}
	dur := e.GetDurationSeconds()
	end := start.Add(time.Duration(dur) * time.Second)
	var notify *int64
	if e.NotifyBeforeMinutes != 0 {
		sec := int64(e.NotifyBeforeMinutes) * 60
		notify = &sec
	}
	return storage.Event{
		ID:           e.GetId(),
		Title:        e.GetTitle(),
		Description:  e.GetDescription(),
		UserID:       e.GetUserId(),
		StartTime:    start.Unix(),
		EndTime:      end.Unix(),
		NotifyBefore: notify,
	}, nil
}

// storageToProtoEvent преобразует storage.Event в pb.Event
func storageToProtoEvent(e storage.Event) *pb.Event {
	start := time.Unix(e.StartTime, 0).Format(time.RFC3339)
	dur := e.EndTime - e.StartTime
	var notify int32
	if e.NotifyBefore != nil {
		notify = int32(*e.NotifyBefore / 60)
	}
	return &pb.Event{
		Id:                  e.ID,
		Title:               e.Title,
		StartTime:           start,
		DurationSeconds:     dur,
		Description:         e.Description,
		UserId:              e.UserID,
		NotifyBeforeMinutes: notify,
	}
}
