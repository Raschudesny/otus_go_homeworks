package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/suite"
)

func TestMemStorage(t *testing.T) {
	suite.Run(t, new(memStorageSuite))
}

type memStorageSuite struct {
	suite.Suite
	storage *MemStorage
	ctx     context.Context
}

func (s *memStorageSuite) SetupTest() {
	s.storage = NewMemStorage()
	s.ctx = context.Background()
}

func (s *memStorageSuite) TestEmpty() {
	s.Require().Equal(int64(0), s.storage.Size(s.ctx))
	events, err := s.storage.FindEventsInInterval(s.ctx, time.Now().AddDate(-1, 0, 0), time.Now().AddDate(1, 0, 0))
	s.Require().NoError(err)
	s.Require().Empty(events)
	events, err = s.storage.FindEventsByID(s.ctx, []string{"1", "2", "3", "4", "5"}...)
	s.Require().NoError(err)
	s.Require().Empty(events)
}

func (s *memStorageSuite) TestDuplicatesNotAdded() {
	var testEvent storage.Event
	err := faker.FakeData(&testEvent)
	s.Require().NoError(err)
	err = s.storage.AddEvent(s.ctx, testEvent)
	s.Require().NoError(err)
	err = s.storage.AddEvent(s.ctx, testEvent)
	s.Require().ErrorIs(err, storage.ErrEventAlreadyExists)
}

func (s *memStorageSuite) TestUpdateEvent() {
	var testEvent storage.Event
	err := faker.FakeData(&testEvent)
	s.Require().NoError(err)
	err = s.storage.AddEvent(s.ctx, testEvent)
	s.Require().NoError(err)

	err = s.storage.UpdateEvent(s.ctx, storage.Event{ID: "not existing id here"})
	s.Require().ErrorIs(err, storage.ErrEventNotFound)

	testEvent.Title = "some new title"
	err = s.storage.UpdateEvent(s.ctx, testEvent)
	s.Require().NoError(err)

	foundEvents, err := s.storage.FindEventsByID(s.ctx, testEvent.ID)
	s.Require().NoError(err)
	s.Require().Equal(1, len(foundEvents))
	s.Require().Equal("some new title", foundEvents[0].Title)
}

func (s *memStorageSuite) TestDeleteEvent() {
	var testEvent storage.Event
	var addedEventsIds []string

	// add first element
	err := faker.FakeData(&testEvent)
	s.Require().NoError(err)
	err = s.storage.AddEvent(s.ctx, testEvent)
	s.Require().NoError(err)
	addedEventsIds = append(addedEventsIds, testEvent.ID)

	// add second element
	err = faker.FakeData(&testEvent)
	s.Require().NoError(err)
	err = s.storage.AddEvent(s.ctx, testEvent)
	s.Require().NoError(err)
	addedEventsIds = append(addedEventsIds, testEvent.ID)

	// try to delete not existing el
	err = s.storage.DeleteEvent(s.ctx, "123456")
	s.Require().ErrorIs(err, storage.ErrEventNotFound)

	for _, eventID := range addedEventsIds {
		err := s.storage.DeleteEvent(s.ctx, eventID)
		s.Require().NoError(err)
	}
	s.Require().Equal(int64(0), s.storage.Size(s.ctx))
}

func (s *memStorageSuite) TestMemStorageBaseUsageConcurrently() {
	var wg sync.WaitGroup
	concurrentUsers := 4
	numOfUserEvents := 5000

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		datesOffsetInDays := numOfUserEvents * i
		datesOffset := time.Now().AddDate(0, 0, datesOffsetInDays)
		go func(offset time.Time) {
			defer wg.Done()
			for j := 0; j < numOfUserEvents; j++ {
				var testEvent storage.Event
				err := faker.FakeData(&testEvent)

				testEvent.StartTime = offset.AddDate(0, 0, j).Add(time.Millisecond)
				// force event duration to be equal to 1 day
				testEvent.EndTime = testEvent.StartTime.AddDate(0, 0, 1)
				s.Require().NoError(err)
				err = s.storage.AddEvent(s.ctx, testEvent)
				s.Require().NoErrorf(err, "testEvent: %+v", testEvent)

				wg.Add(1)
				if j%2 == 0 {
					//
					go func(testEvent storage.Event) {
						defer wg.Done()
						testEvent.Title = "updated event"
						err := s.storage.UpdateEvent(s.ctx, testEvent)
						s.Require().NoError(err)
					}(testEvent)
				} else {
					go func(testEvent storage.Event) {
						defer wg.Done()
						err := s.storage.DeleteEvent(s.ctx, testEvent.ID)
						s.Require().NoError(err)
					}(testEvent)
				}
			}
		}(datesOffset)
	}
	wg.Wait()
	s.Require().Equal(s.storage.Size(s.ctx), int64(concurrentUsers*numOfUserEvents/2))
}

func (s *memStorageSuite) TestFindEventsInInterval() {
	numOfTestEvents := 5
	allAddedEvents := make([]storage.Event, 0, numOfTestEvents)
	for i := 0; i < numOfTestEvents; i++ {
		var testEvent storage.Event
		err := faker.FakeData(&testEvent)
		s.Require().NoError(err)
		testEvent.StartTime = time.Now().AddDate(0, 0, i)
		testEvent.EndTime = testEvent.StartTime.AddDate(0, 0, 1)
		err = s.storage.AddEvent(context.Background(), testEvent)
		s.Require().NoError(err)
		allAddedEvents = append(allAddedEvents, testEvent)
	}
	events, err := s.storage.FindEventsInInterval(context.Background(), time.Now(), time.Now().AddDate(0, 0, numOfTestEvents))
	s.Require().NoError(err)
	s.Require().Equal(len(events), len(allAddedEvents))
	s.Require().ElementsMatch(events, allAddedEvents)
}
