package grpc

import (
	"strconv"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ToStorageTestCase struct {
	input  *pb.Event
	expect *storage.Event
	name   string
}

type FromStorageTestCase struct {
	input  storage.Event
	expect *pb.Event
	name   string
}

func TestMapToPbFormat(t *testing.T) {
	testCases := make([]FromStorageTestCase, 0)
	for i := 0; i < 20; i++ {
		testEvent := storage.Event{}
		err := faker.FakeData(&testEvent)
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond).Local()
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond).Local()
		require.NoError(t, err, "error during fake event generation")

		testCases = append(testCases, FromStorageTestCase{
			input: testEvent,
			expect: &pb.Event{
				Id:          testEvent.ID,
				Title:       testEvent.Title,
				StartTime:   timestamppb.New(testEvent.StartTime),
				EndTime:     timestamppb.New(testEvent.EndTime),
				Description: testEvent.Description,
				OwnerId:     testEvent.OwnerID,
			},
			name: t.Name() + " case number " + strconv.Itoa(i),
		})
	}

	for _, testData := range testCases {
		testData := testData
		t.Run(testData.name, func(t *testing.T) {
			t.Parallel()
			actual := MapToPbFormat(testData.input)
			require.True(t, checkTwoPbEventsEqual(testData.expect, actual))
		})
	}
}

func checkTwoPbEventsEqual(e1 *pb.Event, e2 *pb.Event) bool {
	if e1.Id != e2.Id || e1.Title != e2.Title || e1.Description != e2.Description || e1.OwnerId != e2.OwnerId {
		return false
	}
	if !e1.StartTime.AsTime().Equal(e2.StartTime.AsTime()) {
		return false
	}
	if !e1.EndTime.AsTime().Equal(e2.EndTime.AsTime()) {
		return false
	}
	return true
}

func TestMapToStorageFormat(t *testing.T) {
	testCases := make([]ToStorageTestCase, 0)
	for i := 0; i < 20; i++ {
		testEvent := storage.Event{}
		err := faker.FakeData(&testEvent)
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond).Local()
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond).Local()
		require.NoError(t, err, "error during fake event generation")

		testCases = append(testCases, ToStorageTestCase{
			input:  MapToPbFormat(testEvent),
			expect: &testEvent,
			name:   "Case number " + strconv.Itoa(i),
		})
	}

	for _, testData := range testCases {
		testData := testData
		t.Run(testData.name, func(t *testing.T) {
			t.Parallel()
			actual, err := MapToStorageFormat(testData.input)
			require.NoError(t, err)
			require.True(t, testData.expect.IsEqual(*actual))
		})
	}
}
