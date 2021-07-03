package api

import (
	"context"
	"fmt"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/gofrs/uuid"
)

type EventRepository interface {
	AddEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, eventID string, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	FindEventsInInterval(ctx context.Context, intervalStart, intervalEnd time.Time) ([]storage.Event, error)
	FindEventsByID(ctx context.Context, eventIDs ...string) ([]storage.Event, error)
}

type API interface {
	CreateEvent(ctx context.Context, title string, startTime, endTime time.Time, description, ownerID string) error
	UpdateEvent(ctx context.Context, eventID string, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListWeekEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
	ListMonthEvents(ctx context.Context, date time.Time) ([]storage.Event, error)
}

type api struct {
	cfg  config.Config
	repo EventRepository
}

func New(conf *config.Config, repo EventRepository) API {
	return &api{*conf, repo}
}

func (a *api) CreateEvent(ctx context.Context, title string, startTime, endTime time.Time, description, ownerID string) error {
	uuid4, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("error during generation uuid for event id: %w", err)
	}
	return a.repo.AddEvent(ctx, storage.Event{ID: uuid4.String(), Title: title, StartTime: startTime, EndTime: endTime, Description: description, OwnerID: ownerID})
}

func (a *api) UpdateEvent(ctx context.Context, eventID string, event storage.Event) error {
	return a.repo.UpdateEvent(ctx, eventID, event)
}

func (a *api) DeleteEvent(ctx context.Context, eventID string) error {
	return a.repo.DeleteEvent(ctx, eventID)
}

func (a *api) ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfDay(date)
	intervalEnd := endOfDay(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding events in day interval: %w", err)
	}
	return events, nil
}

func (a *api) ListWeekEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfWeek(date)
	intervalEnd := endOfWeek(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding events in week interval: %w", err)
	}
	return events, nil
}

func (a *api) ListMonthEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfMonth(date)
	intervalEnd := endOfMonth(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding events in month interval: %w", err)
	}
	return events, nil
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func startOfWeek(t time.Time) time.Time {
	return t.Truncate(24 * 7 * time.Hour)
}

func endOfWeek(t time.Time) time.Time {
	return startOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func startOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func endOfMonth(t time.Time) time.Time {
	return startOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}
