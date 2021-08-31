package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
)

type MemStorage struct {
	rw sync.RWMutex
	// TODO this is a disgusting(performance ya know) structure for in memory storage
	// TODO only its advantage is that we can get event by id in complexity of O(1)
	// TODO maybe even a simple slice will perform better because it's sequential
	store map[string]storage.Event
}

func (s *MemStorage) AddEvent(ctx context.Context, event storage.Event) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.store[event.ID]; ok {
		return storage.ErrEventAlreadyExists
	}
	s.store[event.ID] = event
	return nil
}

func (s *MemStorage) UpdateEvent(ctx context.Context, event storage.Event) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.store[event.ID]; !ok {
		return storage.ErrEventNotFound
	}
	s.store[event.ID] = event
	return nil
}

func (s *MemStorage) DeleteEvent(ctx context.Context, eventID string) error {
	s.rw.Lock()
	defer s.rw.Unlock()
	if _, ok := s.store[eventID]; !ok {
		return storage.ErrEventNotFound
	}
	delete(s.store, eventID)
	return nil
}

func (s *MemStorage) FindEventsInInterval(ctx context.Context, intervalStart, intervalEnd time.Time) ([]storage.Event, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var resultEvents []storage.Event
	for _, event := range s.store {
		if isEventInsideTimeInterval(intervalStart, intervalEnd, event.StartTime, event.EndTime) {
			resultEvents = append(resultEvents, event)
		}
	}
	return resultEvents, nil
}

func (s *MemStorage) FindEventsByID(ctx context.Context, eventIDs ...string) ([]storage.Event, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()

	var resultEvents []storage.Event
	for _, eventID := range eventIDs {
		event, ok := s.store[eventID]
		if ok {
			resultEvents = append(resultEvents, event)
		}
	}
	return resultEvents, nil
}

func (s *MemStorage) Size(ctx context.Context) int64 {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return int64(len(s.store))
}

// Checks whether the event inside time interval
// Examples below:
// intervalStart eventStart eventEnd intervalEnd - is OK.
// eventStart intervalStart eventEnd intervalEnd - is OK.
// eventStart intervalStart intervalEnd eventEnd - is OK.
// intervalStart eventStart intervalEnd eventEnd - is OK.
// intervalStart intervalEnd eventStart eventEnd - not OK.
// eventStart eventEnd intervalStart intervalEnd - not OK.
func isEventInsideTimeInterval(intervalStart, intervalEnd, eventStart, eventEnd time.Time) bool {
	// intervalEnd > eventStart && intervalStart < eventEnd
	return intervalEnd.After(eventStart) && intervalStart.Before(eventEnd)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{store: make(map[string]storage.Event)}
}
