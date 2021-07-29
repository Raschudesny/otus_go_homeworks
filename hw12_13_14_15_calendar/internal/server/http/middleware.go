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

		if delegator.responseStatusCode != nil && *delegator.responseStatusCode == http.StatusOK {
			zap.L().Info("request",
				zap.String("IP", r.RemoteAddr),
				zap.String("Time", "["+start.Format(RequestTimeFormat)+"]"),
				zap.String("Method", r.Method),
				zap.String("Path", r.URL.Path),
				zap.String("Version", r.Proto),
				zap.Int("Status", 200),
				zap.Duration("Latency(ms)", latency),
				zap.String("User-Agent", r.UserAgent()),
			)
		}
	})
}

type ResponseWriterDelegator struct {
	delegator          http.ResponseWriter
	responseStatusCode *int
}

func NewResponseWriterDelegator(w http.ResponseWriter) *ResponseWriterDelegator {
	return &ResponseWriterDelegator{w, nil}
}

func (d ResponseWriterDelegator) Header() http.Header {
	return d.delegator.Header()
}

func (d ResponseWriterDelegator) Write(arg []byte) (int, error) {
	return d.delegator.Write(arg)
}

func (d *ResponseWriterDelegator) WriteHeader(statusCode int) {
	d.responseStatusCode = &statusCode
	d.delegator.WriteHeader(statusCode)
}

func (d ResponseWriterDelegator) GetResponseStatusCode() *int {
	return d.responseStatusCode
}
