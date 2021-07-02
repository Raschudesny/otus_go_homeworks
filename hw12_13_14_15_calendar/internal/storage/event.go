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
	ID string `faker:"uuid_hyphenated" db:"id"`
	// Заголовок - короткий текст;
	Title string `faker:"sentence" db:"title"`
	// Дата и время события;
	StartTime time.Time `db:"start_time"`
	// Длительность события (или дата и время окончания);
	EndTime time.Time `db:"end_time"`
	// Описание события - длинный текст, опционально;
	Description string `faker:"paragraph" db:"description"`
	// ID пользователя, владельца события
	OwnerID string `faker:"uuid_hyphenated" db:"owner_id"`
}
