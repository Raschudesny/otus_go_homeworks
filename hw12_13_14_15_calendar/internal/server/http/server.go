package internalhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type API struct {
	server *http.Server
}

func NewHTTPApi(cnf config.HTTPApiConfig, app server.Application) *API {
	service := Service{app}
	router := mux.NewRouter()
	router.HandleFunc("/calendar/add", service.AddEventHandler).Methods("POST")
	router.HandleFunc("/calendar/update", service.UpdateEventHandler).Methods("POST")
	router.HandleFunc("/calendar/delete/{eventId}", service.DeleteEventHandler).Methods("POST")
	router.HandleFunc(
		"/calendar/find/{period:[a-zA-Z]+}/{year:[0-9]{4}}/{month:[0-9]{2}}/{day:[0-9]{2}}",
		service.FindEventsHandler,
	).Methods("GET")

	srv := &http.Server{
		Handler:      loggingMiddleware(router),
		Addr:         net.JoinHostPort("localhost", strconv.Itoa(cnf.Port)),
		ReadTimeout:  time.Duration(cnf.ConnectionTimeout) * time.Second,
		WriteTimeout: time.Duration(cnf.ConnectionTimeout) * time.Second,
	}
	return &API{srv}
}

// Start function is starting http api server on the given port.
// This function is blocking so it must be called in separate goroutine.
// If server start fails, CancelFunc will be called.
func (s *API) Start(cancelFunc context.CancelFunc) {
	zap.L().Info("HTTP server starting...", zap.String("address", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.L().Error("Failed to start http server", zap.Error(err))
		// manually calling server shutdown
		cancelFunc()
	}
}

func (s *API) Stop(ctx context.Context) error {
	zap.L().Info("HTTP server stopping...", zap.String("address", s.server.Addr))
	err := s.server.Shutdown(ctx)
	zap.L().Info("HTTP server stopped")
	return err
}
