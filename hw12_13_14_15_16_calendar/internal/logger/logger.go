// Package logger предоставляет простую реализацию логгера с уровнями логирования.
// Поддерживает уровни: ERROR, WARN, INFO, DEBUG.
package logger

import (
	"fmt"
	"strings"
)

// Level представляет уровень логирования
type Level int

const (
	ErrorLevel Level = iota // 0 - только ошибки
	WarnLevel               // 1 - предупреждения и ошибки
	InfoLevel               // 2 - информация, предупреждения и ошибки
	DebugLevel              // 3 - все сообщения включая отладочные
)

// parseLevel преобразует строковое представление уровня в числовое.
// Поддерживаемые значения: "error", "warn", "info", "debug".
// По умолчанию возвращает InfoLevel.
func parseLevel(level string) Level {
	switch strings.ToLower(level) {
	case "error":
		return ErrorLevel
	case "warn":
		return WarnLevel
	case "info":
		return InfoLevel
	case "debug":
		return DebugLevel
	default:
		return InfoLevel
	}
}

// Logger представляет логгер с настраиваемым уровнем логирования
type Logger struct {
	level Level // текущий уровень логирования
}

// New создает новый экземпляр логгера с указанным уровнем
func New(level string) *Logger {
	return &Logger{level: parseLevel(level)}
}

// Error выводит сообщение об ошибке, если текущий уровень >= ErrorLevel
func (l *Logger) Error(msg string) {
	if l.level >= ErrorLevel {
		fmt.Println("[ERROR]", msg)
	}
}

// Warn выводит предупреждение, если текущий уровень >= WarnLevel
func (l *Logger) Warn(msg string) {
	if l.level >= WarnLevel {
		fmt.Println("[WARN]", msg)
	}
}

// Info выводит информационное сообщение, если текущий уровень >= InfoLevel
func (l *Logger) Info(msg string) {
	if l.level >= InfoLevel {
		fmt.Println("[INFO]", msg)
	}
}

// Debug выводит отладочное сообщение, если текущий уровень >= DebugLevel
func (l *Logger) Debug(msg string) {
	if l.level >= DebugLevel {
		fmt.Println("[DEBUG]", msg)
	}
}
