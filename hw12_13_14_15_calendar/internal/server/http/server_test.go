package internalhttp

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/stretchr/testify/require"
)

func TestServerBasic(t *testing.T) {
	wg := sync.WaitGroup{}
	serverContext, cancelFunc := context.WithCancel(context.Background())
	server := NewServer(&config.Config{
		Logger:  config.LoggerConfig{},
		Storage: config.StorageConfig{},
		API:     config.APIConfig{Port: 9999},
	}, nil)

	wg.Add(2)
	go func(serverStop context.CancelFunc) {
		defer wg.Done()
		// defer serverStop call because we need server to stop anyway nevermind failed test or succeeded
		defer serverStop()
		timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()

		req, err := http.NewRequestWithContext(timeout, http.MethodGet, "http://localhost:9999/hello", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() {
			err := resp.Body.Close()
			require.NoError(t, err)
		}()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "Hello World!", string(body))
	}(cancelFunc)
	go func() {
		defer wg.Done()
		<-serverContext.Done()
		timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		err := server.Stop(timeout)
		require.NoError(t, err)
	}()

	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		require.NoError(t, err)
	}
	wg.Wait()
}
