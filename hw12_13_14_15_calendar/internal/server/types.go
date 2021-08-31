package server

import (
	"context"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
)

//go:generate mockgen --build_flags=--mod=mod -destination=./mock_types.go -package=server . Application
type Application interface {
	CreateEvent(ctx context.Context, title string, startTime, endTime time.Time, description, ownerID string) (storage.Event, error)
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListWeekEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListMonthEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
}
