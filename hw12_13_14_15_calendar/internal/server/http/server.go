package internalhttp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/api"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
	api    api.API
}

func HelloWorldResource(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := r.Body.Close(); err != nil {
			zap.L().Error("unable to close response body for request", zap.Error(err))
		}
	}()
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello World!")); err != nil {
		zap.L().Error("http write error", zap.Error(err))
	}
}

func NewServer(conf *config.Config, api api.API) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", HelloWorldResource)

	server := &http.Server{
		Handler:      loggingMiddleware(mux),
		Addr:         ":" + strconv.Itoa(conf.API.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	return &Server{server, api}
}

func (s *Server) Start() error {
	zap.L().Info("Server starting...", zap.String("address", s.server.Addr))
	return s.server.ListenAndServe()
}

func (s *Server) Stop(timeoutCtx context.Context) error {
	zap.L().Info("Server stopping...", zap.String("address", s.server.Addr))
	return s.server.Shutdown(timeoutCtx)
}
