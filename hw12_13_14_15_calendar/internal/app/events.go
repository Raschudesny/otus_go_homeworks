package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/gofrs/uuid"
)

type EventRepository interface {
	AddEvent(ctx context.Context, event storage.Event) error
	UpdateEvent(ctx context.Context, event storage.Event) error
	DeleteEvent(ctx context.Context, eventID string) error
	FindEventsInInterval(ctx context.Context, intervalStart, intervalEnd time.Time) ([]storage.Event, error)
	FindEventsByID(ctx context.Context, eventIDs ...string) ([]storage.Event, error)
}

type EventsService struct {
	repo EventRepository
}

func New(repo EventRepository) *EventsService {
	return &EventsService{repo}
}

func (a *EventsService) CreateEvent(ctx context.Context, title string, startTime, endTime time.Time, description, ownerID string) (storage.Event, error) {
	uuid4, err := uuid.NewV4()
	if err != nil {
		return storage.Event{}, fmt.Errorf("error during generation uuid for event id: %w", err)
	}
	event := storage.Event{ID: uuid4.String(), Title: title, StartTime: startTime, EndTime: endTime, Description: description, OwnerID: ownerID}
	err = a.repo.AddEvent(ctx, event)
	if err != nil {
		return storage.Event{}, fmt.Errorf("error during creating event: %w", err)
	}
	return event, nil
}

func (a *EventsService) UpdateEvent(ctx context.Context, event storage.Event) error {
	return a.repo.UpdateEvent(ctx, event)
}

func (a *EventsService) DeleteEvent(ctx context.Context, eventID string) error {
	return a.repo.DeleteEvent(ctx, eventID)
}

func (a *EventsService) ListDayEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfDay(date)
	intervalEnd := endOfDay(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding app in day interval: %w", err)
	}
	return events, nil
}

func (a *EventsService) ListWeekEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfWeek(date)
	intervalEnd := endOfWeek(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding app in week interval: %w", err)
	}
	return events, nil
}

func (a *EventsService) ListMonthEvents(ctx context.Context, date time.Time) ([]storage.Event, error) {
	intervalStart := startOfMonth(date)
	intervalEnd := endOfMonth(date)
	events, err := a.repo.FindEventsInInterval(ctx, intervalStart, intervalEnd)
	if err != nil {
		return nil, fmt.Errorf("errod during finding app in month interval: %w", err)
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
