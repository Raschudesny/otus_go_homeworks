package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/require"
)

// TODO переписать тесты на сьюты а то это какой-то кошмар

func TestEmpty(t *testing.T) {
	memStorage := New()
	testContext := context.Background()
	require.Equal(t, int64(0), memStorage.Size(testContext))
	events, err := memStorage.FindEventsInInterval(testContext, time.Now().AddDate(-1, 0, 0), time.Now().AddDate(1, 0, 0))
	require.NoError(t, err)
	require.Empty(t, events)
	events, err = memStorage.FindEventsByID(testContext, []string{"1", "2", "3", "4", "5"}...)
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestDuplicatesNotAdded(t *testing.T) {
	memStorage := New()
	testContext := context.Background()
	var testEvent storage.Event
	err := faker.FakeData(&testEvent)
	require.NoError(t, err)
	err = memStorage.AddEvent(testContext, testEvent)
	require.NoError(t, err)
	err = memStorage.AddEvent(testContext, testEvent)
	require.ErrorIs(t, err, storage.ErrEventAlreadyExists)
}

func TestUpdateEvent(t *testing.T) {
	memStorage := New()
	testContext := context.Background()
	var testEvent storage.Event
	err := faker.FakeData(&testEvent)
	require.NoError(t, err)
	err = memStorage.AddEvent(testContext, testEvent)
	require.NoError(t, err)

	err = memStorage.UpdateEvent(testContext, "not existing id here", testEvent)
	require.ErrorIs(t, err, storage.ErrEventNotFound)

	testEvent.Title = "some new title"
	err = memStorage.UpdateEvent(testContext, testEvent.ID, testEvent)
	require.NoError(t, err)

	foundEvents, err := memStorage.FindEventsByID(testContext, testEvent.ID)
	require.NoError(t, err)
	require.Equal(t, 1, len(foundEvents))
	require.Equal(t, "some new title", foundEvents[0].Title)
}

func TestDeleteEvent(t *testing.T) {
	memStorage := New()
	testContext := context.Background()
	var testEvent storage.Event

	var addedEventsIds []string

	// add first element
	err := faker.FakeData(&testEvent)
	require.NoError(t, err)
	err = memStorage.AddEvent(testContext, testEvent)
	require.NoError(t, err)
	addedEventsIds = append(addedEventsIds, testEvent.ID)

	// add second element
	err = faker.FakeData(&testEvent)
	require.NoError(t, err)
	err = memStorage.AddEvent(testContext, testEvent)
	require.NoError(t, err)
	addedEventsIds = append(addedEventsIds, testEvent.ID)

	// try delete not exist el
	err = memStorage.DeleteEvent(testContext, "123456")
	require.ErrorIs(t, err, storage.ErrEventNotFound)

	for _, eventID := range addedEventsIds {
		err := memStorage.DeleteEvent(testContext, eventID)
		require.NoError(t, err)
	}
	require.Equal(t, int64(0), memStorage.Size(testContext))
}

func TestMemStorageBaseUsageConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	memStorage := New()
	concurrentUsers := 4
	numOfUserEvents := 5000
	background := context.Background()

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
				require.NoError(t, err)
				err = memStorage.AddEvent(background, testEvent)
				require.NoErrorf(t, err, "testEvent: %+v", testEvent)

				wg.Add(1)
				if j%2 == 0 {
					//
					go func(testEvent storage.Event) {
						defer wg.Done()
						testEvent.Title = "updated event"
						err := memStorage.UpdateEvent(background, testEvent.ID, testEvent)
						require.NoError(t, err)
					}(testEvent)
				} else {
					go func(testEvent storage.Event) {
						defer wg.Done()
						err := memStorage.DeleteEvent(background, testEvent.ID)
						require.NoError(t, err)
					}(testEvent)
				}
			}
		}(datesOffset)
	}
	wg.Wait()
	require.Equal(t, memStorage.Size(background), int64(concurrentUsers*numOfUserEvents/2))
}

func TestFindEventsInInterval(t *testing.T) {
	memStorage := New()
	numOfTestEvents := 5
	allAddedEvents := make([]storage.Event, 0, numOfTestEvents)
	for i := 0; i < numOfTestEvents; i++ {
		var testEvent storage.Event
		err := faker.FakeData(&testEvent)
		require.NoError(t, err)
		testEvent.StartTime = time.Now().AddDate(0, 0, i)
		testEvent.EndTime = testEvent.StartTime.AddDate(0, 0, 1)
		err = memStorage.AddEvent(context.Background(), testEvent)
		require.NoError(t, err)
		allAddedEvents = append(allAddedEvents, testEvent)
	}
	events, err := memStorage.FindEventsInInterval(context.Background(), time.Now(), time.Now().AddDate(0, 0, numOfTestEvents))
	require.NoError(t, err)
	require.Equal(t, len(events), len(allAddedEvents))
	require.ElementsMatch(t, events, allAddedEvents)
}
