// Package storage содержит общие типы данных для работы с хранилищем событий
package storage

// Event представляет событие в календаре.
// Содержит всю необходимую информацию о событии пользователя.
type Event struct {
	ID           string // уникальный идентификатор события (UUID)
	Title        string // заголовок события
	Description  string // описание события
	UserID       string // идентификатор пользователя, владельца события
	StartTime    int64  // время начала события (Unix timestamp)
	EndTime      int64  // время окончания события (Unix timestamp)
	NotifyBefore *int64 // количество секунд до события для уведомления (опционально)
}
