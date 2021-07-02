package sqlstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DBStorage struct {
	db *sqlx.DB
}

func New() *DBStorage {
	return &DBStorage{}
}

func (s *DBStorage) Connect(ctx context.Context, dsn string) (err error) {
	s.db, err = sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}
	return s.db.PingContext(ctx)
}

func (s *DBStorage) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error during db connection pool closing: %w", err)
	}
	return nil
}

func (s *DBStorage) AddEvent(ctx context.Context, event storage.Event) error {
	duplicateEvents, err := s.FindEventsByID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("error during duplicate check: %w", err)
	}
	if len(duplicateEvents) > 0 {
		return storage.ErrEventAlreadyExists
	}

	_, err = s.db.NamedExecContext(ctx, "INSERT INTO events (id, title, start_time, end_time, description, owner_id) VALUES (:id, :title, :start_time, :end_time, :description, :owner_id)", &event)
	if err != nil {
		return fmt.Errorf("error during add event sql execution: %w", err)
	}
	return nil
}

func (s *DBStorage) UpdateEvent(ctx context.Context, eventID string, event storage.Event) error {
	res, err := s.db.NamedExecContext(ctx, "UPDATE events SET title=:title, start_time=:start_time, end_time=:end_time, description=:description, owner_id=:owner_id WHERE id=:id", &event)
	if err != nil {
		return fmt.Errorf("error during updating event: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by update checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *DBStorage) DeleteEvent(ctx context.Context, eventID string) error {
	res, err := s.db.NamedExecContext(ctx, "DELETE FROM events WHERE id=:id", map[string]interface{}{
		"id": eventID,
	})
	if err != nil {
		return fmt.Errorf("error during deleting event: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error during rows affected by update checking: %w", err)
	}
	if affected == 0 {
		return storage.ErrEventNotFound
	}
	return nil
}

func (s *DBStorage) FindEventsInInterval(ctx context.Context, intervalStart, intervalEnd time.Time) ([]storage.Event, error) {
	sql := "select * from events where start_time < :intervalEnd AND end_time > :intervalStart"
	rows, err := s.db.NamedQueryContext(ctx, sql, map[string]interface{}{
		"intervalStart": intervalStart,
		"intervalEnd":   intervalEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("sql execution error: %w", err)
	}
	defer func() {
		err = rows.Close()
		zap.L().Error("error closing sql rows", zap.Error(err))
	}()

	var result []storage.Event
	var event storage.Event
	for rows.Next() {
		if err := rows.StructScan(&event); err != nil {
			return nil, fmt.Errorf("sql result event parsing error: %w", err)
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sql result event parsing error: %w", err)
	}
	return result, nil
}

func (s *DBStorage) FindEventsByID(ctx context.Context, eventIDs ...string) ([]storage.Event, error) {
	var result []storage.Event
	if len(eventIDs) == 0 {
		return nil, nil
	}
	if len(eventIDs) == 1 {
		sql := "select * from events where id = :id"
		rows, err := s.db.NamedQueryContext(ctx, sql, map[string]interface{}{
			"id": eventIDs[0],
		})
		if err != nil {
			return nil, fmt.Errorf("sql execution error: %w", err)
		}
		defer func() {
			err = rows.Close()
			zap.L().Error("error closing sql rows", zap.Error(err))
		}()

		var event storage.Event
		for rows.Next() {
			if err := rows.StructScan(&event); err != nil {
				return nil, fmt.Errorf("sql result event parsing error: %w", err)
			}
			result = append(result, event)
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("sql result event parsing error: %w", err)
		}

		return result, nil
	}

	query := "select * from events where id in (?)"
	query, args, err := sqlx.In(query, eventIDs)
	if err != nil {
		return nil, fmt.Errorf("error during preparing sql: %w", err)
	}
	resultQuery := s.db.Rebind(query)

	rows, err := s.db.QueryxContext(ctx, resultQuery, args)
	defer func() {
		err = rows.Close()
		zap.L().Error("error closing sql rows", zap.Error(err))
	}()
	if err != nil {
		return nil, fmt.Errorf("sql execution error: %w", err)
	}

	var event storage.Event
	for rows.Next() {
		if err := rows.StructScan(&event); err != nil {
			return nil, fmt.Errorf("sql result event parsing error: %w", err)
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sql result event parsing error: %w", err)
	}
	return result, nil
}
