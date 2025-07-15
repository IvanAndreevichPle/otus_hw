package internalhttp

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// loggingMiddleware создает middleware для логирования HTTP-запросов.
// Логирует IP клиента, время запроса, метод, путь, версию HTTP, код ответа,
// время обработки запроса и User-Agent.
func loggingMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Засекаем время начала обработки запроса
			start := time.Now()

			// Оборачиваем ResponseWriter для получения статус-кода
			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, r)

			// Вычисляем время обработки запроса
			latency := time.Since(start)

			// Определяем реальный IP клиента
			ip := r.RemoteAddr
			if ipHeader := r.Header.Get("X-Real-IP"); ipHeader != "" {
				// Используем X-Real-IP заголовок (обычно устанавливается nginx)
				ip = ipHeader
			} else if ipHeader := r.Header.Get("X-Forwarded-For"); ipHeader != "" {
				// Используем X-Forwarded-For заголовок (первый IP в списке)
				ip = strings.Split(ipHeader, ",")[0]
			}

			// Форматируем временную метку в стандартном формате
			timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")

			// Получаем User-Agent
			ua := r.UserAgent()

			// Формируем строку лога в формате Apache/Nginx
			logLine := fmt.Sprintf("%s [%s] %s %s %s %d %d \"%s\"",
				ip,
				timestamp,
				r.Method,
				r.RequestURI,
				r.Proto,
				rw.statusCode,
				latency.Milliseconds(),
				ua,
			)
			logger.Info(logLine)
		})
	}
}

// responseWriter оборачивает http.ResponseWriter для получения статус-кода ответа
type responseWriter struct {
	http.ResponseWriter
	statusCode int // сохраняем статус-код для логирования
}

// WriteHeader переопределяет метод для сохранения статус-кода
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
