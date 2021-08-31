package grpc

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/app"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/grpc/pb"
	memorystorage "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage/memory"
	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCTestSuite struct {
	suite.Suite

	grpcServer     *grpc.Server
	grpcClientConn *grpc.ClientConn
	ctx            context.Context
	cancelFunc     context.CancelFunc
}

func TestGrpc(t *testing.T) {
	suite.Run(t, &GRPCTestSuite{})
}

func (s *GRPCTestSuite) SetupSuite() {
	lsnStub := bufconn.Listen(1024 * 1024)

	// starting grpc server
	s.grpcServer = grpc.NewServer(grpc.ConnectionTimeout(5 * time.Second))
	pb.RegisterCalendarServiceServer(s.grpcServer, &CalendarService{app: app.New(memorystorage.NewMemStorage())})
	go func() {
		if err := s.grpcServer.Serve(lsnStub); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatal("error during grpc test server stating: ", err)
		}
		// BANNED ERROR CHECK VIA TEST HERE
		// because instead we are getting data race in this case ??? (do you see it? ... but it's here)
		// s.Require().NoError(err, "grpc server failed starting")
	}()

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	s.ctx = ctx
	s.cancelFunc = cancelFunc

	conn, err := grpc.DialContext(ctx,
		"",
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lsnStub.Dial()
		}))
	s.Require().NoError(err)

	s.grpcClientConn = conn
}

func (s *GRPCTestSuite) TearDownSuite() {
	defer s.cancelFunc()
	if err := s.grpcClientConn.Close(); err != nil {
		s.T().Log("error during connection closing: ", err)
	}
	s.grpcServer.Stop()
}

func (s *GRPCTestSuite) TestAddEvent() {
	client := pb.NewCalendarServiceClient(s.grpcClientConn)

	data := pb.AddEventRequest_CreateEventData{
		Title:       faker.Sentence(),
		StartTime:   timestamppb.New(time.Now().Truncate(time.Nanosecond).Local()),
		EndTime:     timestamppb.New(time.Now().AddDate(0, 0, 1).Truncate(time.Nanosecond).Local()),
		Description: faker.Paragraph(),
		OwnerId:     faker.UUIDHyphenated(),
	}

	resp, err := client.AddEvent(s.ctx, &pb.AddEventRequest{
		CreateEventData: &data,
	})
	s.Require().NoError(err)
	s.Require().Equal(data.Title, resp.GetEvent().GetTitle())
	s.Require().Equal(data.Description, resp.GetEvent().GetDescription())
	s.Require().True(data.StartTime.AsTime().Equal(resp.GetEvent().GetStartTime().AsTime()))
	s.Require().True(data.EndTime.AsTime().Equal(resp.GetEvent().GetEndTime().AsTime()))
	s.Require().Equal(data.OwnerId, resp.GetEvent().GetOwnerId())
}

func (s *GRPCTestSuite) TestUpdateEvent() {
	client := pb.NewCalendarServiceClient(s.grpcClientConn)

	// adding event
	data := pb.AddEventRequest_CreateEventData{
		Title:       faker.Sentence(),
		StartTime:   timestamppb.New(time.Now().Truncate(time.Nanosecond).Local()),
		EndTime:     timestamppb.New(time.Now().AddDate(0, 0, 1).Truncate(time.Nanosecond).Local()),
		Description: faker.Paragraph(),
		OwnerId:     faker.UUIDHyphenated(),
	}
	resp, err := client.AddEvent(s.ctx, &pb.AddEventRequest{
		CreateEventData: &data,
	})
	s.Require().NoError(err)

	// updating event
	event := resp.GetEvent()
	event.Title = "updated"
	updateResp, err := client.UpdateEvent(s.ctx, &pb.UpdateEventRequest{Event: event})
	s.Require().NoError(err)
	expected, err := MapToStorageFormat(event)
	s.Require().NoError(err)
	actual, err := MapToStorageFormat(updateResp.Event)
	s.Require().NoError(err)
	s.Require().True(expected.IsEqual(*actual))
}

func (s *GRPCTestSuite) TestDeleteEvent() {
	client := pb.NewCalendarServiceClient(s.grpcClientConn)

	// adding event
	data := pb.AddEventRequest_CreateEventData{
		Title:       faker.Sentence(),
		StartTime:   timestamppb.New(time.Now().Truncate(time.Nanosecond).Local()),
		EndTime:     timestamppb.New(time.Now().AddDate(0, 0, 1).Truncate(time.Nanosecond).Local()),
		Description: faker.Paragraph(),
		OwnerId:     faker.UUIDHyphenated(),
	}
	resp, err := client.AddEvent(s.ctx, &pb.AddEventRequest{
		CreateEventData: &data,
	})
	s.Require().NoError(err)

	// deleting event
	eventID := resp.GetEvent().GetId()
	_, err = client.DeleteEvent(s.ctx, &pb.DeleteEventRequest{EventId: eventID})
	s.Require().NoError(err)
}

func (s *GRPCTestSuite) TestFindEvents() {
	client := pb.NewCalendarServiceClient(s.grpcClientConn)

	// adding event
	t := time.Now().Truncate(time.Nanosecond).Local()
	data := pb.AddEventRequest_CreateEventData{
		Title:       faker.Sentence(),
		StartTime:   timestamppb.New(t),
		EndTime:     timestamppb.New(t.AddDate(0, 0, 1)),
		Description: faker.Paragraph(),
		OwnerId:     faker.UUIDHyphenated(),
	}
	resp, err := client.AddEvent(s.ctx, &pb.AddEventRequest{
		CreateEventData: &data,
	})
	s.Require().NoError(err)

	// finding day events
	findDayResp, err := client.FindDayEvents(s.ctx, &pb.FindDayEventsRequest{Day: timestamppb.New(t)})
	s.Require().NoError(err)
	s.Require().True(PbEventsContains(findDayResp.Events, resp.GetEvent()))

	// finding week events
	findWeekResp, err := client.FindWeekEvents(s.ctx, &pb.FindWeekEventsRequest{Week: timestamppb.New(t)})
	s.Require().NoError(err)
	s.Require().True(PbEventsContains(findWeekResp.Events, resp.GetEvent()))

	// finding month events
	findMonthResp, err := client.FindWeekEvents(s.ctx, &pb.FindWeekEventsRequest{Week: timestamppb.New(t)})
	s.Require().NoError(err)
	s.Require().True(PbEventsContains(findMonthResp.Events, resp.GetEvent()))
}

func PbEventsContains(events []*pb.Event, event *pb.Event) bool {
	e2, err := MapToStorageFormat(event)
	if err != nil {
		return false
	}

	for _, v := range events {
		e1, err := MapToStorageFormat(v)
		if err != nil {
			return false
		}
		if e1.IsEqual(*e2) {
			return true
		}
	}
	return false
}
