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
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond)
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond)
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
			require.Equal(t, testData.expect, actual)
		})
	}
}

func TestMapToStorageFormat(t *testing.T) {
	testCases := make([]ToStorageTestCase, 0)
	for i := 0; i < 20; i++ {
		testEvent := storage.Event{}
		err := faker.FakeData(&testEvent)
		testEvent.StartTime = testEvent.StartTime.Truncate(time.Nanosecond)
		testEvent.EndTime = testEvent.EndTime.Truncate(time.Nanosecond)
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
			require.Equal(t, testData.expect, actual)
		})
	}
}
