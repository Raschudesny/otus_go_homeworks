package internalhttp

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var RequestTimeFormat string = "25/Feb/2020:19:11:24 +0600"

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		delegator := NewResponseWriterDelegator(w)
		start := time.Now()
		next.ServeHTTP(delegator, r)
		latency := time.Since(start)

		zap.L().Info("request",
			zap.String("IP", r.RemoteAddr),
			zap.Time("Time", start),
			zap.String("Method", r.Method),
			zap.String("Path", r.URL.Path),
			zap.String("Version", r.Proto),
			zap.Int("Status", delegator.responseStatusCode),
			zap.Duration("Latency(ms)", latency),
			zap.String("User-Agent", r.UserAgent()),
		)
	})
}

type ResponseWriterDelegator struct {
	http.ResponseWriter
	responseStatusCode int
}

func NewResponseWriterDelegator(w http.ResponseWriter) *ResponseWriterDelegator {
	return &ResponseWriterDelegator{w, http.StatusOK}
}

func (d *ResponseWriterDelegator) WriteHeader(statusCode int) {
	d.responseStatusCode = statusCode
	d.ResponseWriter.WriteHeader(statusCode)
}
