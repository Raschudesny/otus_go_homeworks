package storage

import (
	"errors"
	"time"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
)

type Event struct {
	// ID - уникальный идентификатор события (можно воспользоваться UUID);
	ID string `faker:"uuid_hyphenated" db:"id" json:"id"`
	// Заголовок - короткий текст;
	Title string `faker:"sentence" db:"title" json:"title"`
	// Дата и время события;
	StartTime time.Time `db:"start_time" json:"start_time"`
	// Длительность события (или дата и время окончания);
	EndTime time.Time `db:"end_time" json:"end_time"`
	// Описание события - длинный текст, опционально;
	Description string `faker:"paragraph" db:"description" json:"description"`
	// ID пользователя, владельца события
	OwnerID string `faker:"uuid_hyphenated" db:"owner_id" json:"owner_id"`
}

// IsEqual - check two events is equal, this function is mostly used in tests.
func (e1 Event) IsEqual(e2 Event) bool {
	if e1.ID != e2.ID || e1.Title != e2.Title || e1.Description != e2.Description || e1.OwnerID != e2.OwnerID {
		return false
	}
	return e1.StartTime.Equal(e2.StartTime) && e1.EndTime.Equal(e2.EndTime)
}
