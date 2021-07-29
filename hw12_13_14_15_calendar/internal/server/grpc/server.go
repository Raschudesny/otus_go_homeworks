package grpc

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/grpc/pb"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate protoc --proto_path=../../../api --go_out=pb --go-grpc_out=pb ../../../api/calendar_service.proto
var _ pb.CalendarServiceServer = (*CalendarService)(nil)

type CalendarService struct {
	pb.UnimplementedCalendarServiceServer
	app server.Application
}

func (c *CalendarService) AddEvent(ctx context.Context, request *pb.AddEventRequest) (*pb.AddEventResponse, error) {
	if request.GetCreateEventData() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no event data provided")
	}
	eventData := request.GetCreateEventData()

	if err := eventData.StartTime.CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "start time timestamp is not valid: %s", err)
	}
	if err := eventData.EndTime.CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "end time timestamp is not valid: %s", err)
	}

	event, err := c.app.CreateEvent(
		ctx,
		eventData.Title,
		eventData.StartTime.AsTime(),
		eventData.EndTime.AsTime(),
		eventData.Description,
		eventData.OwnerId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to create event: %s", err)
	}

	return &pb.AddEventResponse{Event: MapToPbFormat(event)}, nil
}

func (c *CalendarService) UpdateEvent(ctx context.Context, request *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	event, err := MapToStorageFormat(request.GetEvent())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %s", err)
	}
	err = c.app.UpdateEvent(ctx, *event)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to update event: %s", err)
	}
	return &pb.UpdateEventResponse{Event: request.GetEvent()}, nil
}

func (c *CalendarService) DeleteEvent(ctx context.Context, request *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if request.GetEventId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "event id validation error: %s", ErrValueIsEmpty)
	}
	eventID := request.GetEventId()
	err := c.app.DeleteEvent(ctx, eventID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to delete event: %s", err)
	}
	return new(pb.DeleteEventResponse), nil
}

func (c *CalendarService) FindDayEvents(ctx context.Context, request *pb.FindDayEventsRequest) (*pb.FindDayEventsResponse, error) {
	if request.GetDay() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "day validation error: %s", ErrValueIsNil)
	}
	if err := request.GetDay().CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "day validation error: %s", err)
	}
	day := request.GetDay().AsTime()
	events, err := c.app.ListDayEvents(ctx, day)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to find day events: %s", err)
	}

	return &pb.FindDayEventsResponse{Events: MapSliceToPbFormat(events)}, nil
}

func (c *CalendarService) FindWeekEvents(ctx context.Context, request *pb.FindWeekEventsRequest) (*pb.FindWeekEventsResponse, error) {
	if request.GetWeek() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "week validation error: %s", ErrValueIsNil)
	}
	if err := request.GetWeek().CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "week validation error: %s", err)
	}
	week := request.GetWeek().AsTime()
	events, err := c.app.ListWeekEvents(ctx, week)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to find week events: %s", err)
	}

	return &pb.FindWeekEventsResponse{Events: MapSliceToPbFormat(events)}, nil
}

func (c *CalendarService) FindMonthEvents(ctx context.Context, request *pb.FindMonthEventsRequest) (*pb.FindMonthEventsResponse, error) {
	if request.GetMonth() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "month validation error: %s", ErrValueIsNil)
	}
	if err := request.GetMonth().CheckValid(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "month validation error: %s", err)
	}
	month := request.GetMonth().AsTime()
	events, err := c.app.ListMonthEvents(ctx, month)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to find month events: %s", err)
	}

	return &pb.FindMonthEventsResponse{Events: MapSliceToPbFormat(events)}, nil
}

type API struct {
	Server *grpc.Server
	port   int
}

// Start function is starting grpc api server on the given port.
// This function is blocking so it must be called in separate goroutine.
// If server start fails, CancelFunc will be called.
func (a API) Start(cancelFunc context.CancelFunc) {
	// manually calling server shutdown
	defer cancelFunc()

	zap.L().Info("GRPC server starting...", zap.String("address", net.JoinHostPort("localhost", strconv.Itoa(a.port))))
	lsn, err := net.Listen("tcp", net.JoinHostPort("localhost", strconv.Itoa(a.port)))
	if err != nil {
		zap.L().Error("Failed to start grpc server", zap.Error(err))
		return
	}
	if err := a.Server.Serve(lsn); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		zap.L().Error("Failed to start grpc server", zap.Error(err))
		return
	}
}

func (a API) Stop() {
	zap.L().Info("GRPC server stopping...", zap.String("address", net.JoinHostPort("localhost", strconv.Itoa(a.port))))
	a.Server.GracefulStop()
	zap.L().Info("GRPC server stopped")
}

func NewGRPCApi(cfg config.GRPCApiConfig, app server.Application) *API {
	srv := grpc.NewServer(
		grpc.ConnectionTimeout(time.Duration(cfg.ConnectionTimeout)*time.Second),
		grpc.UnaryInterceptor(grpc_zap.UnaryServerInterceptor(zap.L())),
		grpc.StreamInterceptor(grpc_zap.StreamServerInterceptor(zap.L())),
	)
	pb.RegisterCalendarServiceServer(srv, &CalendarService{app: app})
	return &API{srv, cfg.Port}
}
