package grpc

import (
	"errors"
	"fmt"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrValueIsNil   = errors.New("value is nil")
	ErrValueIsEmpty = errors.New("value is empty")
)

func MapToStorageFormat(event *pb.Event) (*storage.Event, error) {
	if err := ValidatePbEvent(event); err != nil {
		return nil, fmt.Errorf("failed to map from protobuf object, format error: %w", err)
	}

	return &storage.Event{
		ID:          event.Id,
		Title:       event.Title,
		StartTime:   event.StartTime.AsTime().Local(),
		EndTime:     event.EndTime.AsTime().Local(),
		Description: event.Description,
		OwnerID:     event.OwnerId,
	}, nil
}

func MapToPbFormat(event storage.Event) *pb.Event {
	return &pb.Event{
		Id:          event.ID,
		Title:       event.Title,
		StartTime:   timestamppb.New(event.StartTime),
		EndTime:     timestamppb.New(event.EndTime),
		Description: event.Description,
		OwnerId:     event.OwnerID,
	}
}

func MapSliceToPbFormat(events []storage.Event) []*pb.Event {
	res := make([]*pb.Event, 0, len(events))
	for _, v := range events {
		res = append(res, MapToPbFormat(v))
	}
	return res
}

func ValidatePbEvent(event *pb.Event) error {
	err := func(event *pb.Event) error {
		if event == nil {
			return ErrValueIsNil
		}
		if event.Id == "" {
			return fmt.Errorf("id validation err: %w", ErrValueIsEmpty)
		}
		if event.Description == "" {
			return fmt.Errorf("description validation err: %w", ErrValueIsEmpty)
		}
		if event.Title == "" {
			return fmt.Errorf("title validation err: %w", ErrValueIsEmpty)
		}
		if event.OwnerId == "" {
			return fmt.Errorf("owner id validation err: %w", ErrValueIsEmpty)
		}
		if err := event.StartTime.CheckValid(); err != nil {
			return fmt.Errorf("start time validation err: %w", err)
		}
		if err := event.EndTime.CheckValid(); err != nil {
			return fmt.Errorf("end time validation err: %w", err)
		}
		return nil
	}(event)
	if err != nil {
		return fmt.Errorf("pb event validation error: %w", err)
	}
	return nil
}
