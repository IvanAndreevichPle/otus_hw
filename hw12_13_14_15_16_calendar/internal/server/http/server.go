// Package internalhttp предоставляет HTTP-сервер для приложения календаря.
// Включает middleware для логирования запросов и graceful shutdown.
package internalhttp

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server представляет HTTP-сервер приложения календаря
type Server struct {
	logger  Logger       // логгер для записи событий сервера
	app     Application  // интерфейс к бизнес-логике приложения
	httpSrv *http.Server // встроенный HTTP-сервер Go
}

// Logger определяет интерфейс для логирования
type Logger interface {
	Info(msg string)
	Error(msg string)
	Warn(msg string)
	Debug(msg string)
}

// Application определяет интерфейс к бизнес-логике приложения
// Пока не используется, но может быть расширено для обработки запросов
type Application interface {
	// пока не используется, но может быть расширено для бизнес-логики
}

// NewServer создает новый HTTP-сервер с настроенными маршрутами и middleware
func NewServer(logger Logger, app Application, host string, port int) *Server {
	// Создаем мультиплексор для маршрутизации запросов
	mux := http.NewServeMux()

	// Регистрируем тестовый эндпоинт
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "hello world")
	})

	// Оборачиваем мультиплексор в middleware для логирования
	h := loggingMiddleware(logger)(mux)

	// Формируем адрес для сервера
	addr := fmt.Sprintf("%s:%d", host, port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	return &Server{
		logger:  logger,
		app:     app,
		httpSrv: httpSrv,
	}
}

// Start запускает HTTP-сервер и начинает прослушивание входящих запросов
func (s *Server) Start(ctx context.Context) error {
	// Горутина для graceful shutdown при отмене контекста
	go func() {
		<-ctx.Done()
		_ = s.Stop(context.Background())
	}()
	return s.httpSrv.ListenAndServe()
}

// Stop останавливает HTTP-сервер с graceful shutdown
// Дает серверу время на завершение обработки текущих запросов
func (s *Server) Stop(ctx context.Context) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return s.httpSrv.Shutdown(ctxTimeout)
}
