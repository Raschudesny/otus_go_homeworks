package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/api"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var configFilePath string

var ServerShutdownTimeout = time.Second * 3

func init() {
	pflag.StringVarP(&configFilePath, "config", "c", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	// forcing only one Fatal in an app
	if err := mainImpl(); err != nil {
		log.Fatal(err)
	}
}

func mainImpl() error {
	pflag.Parse()
	for _, arg := range pflag.Args() {
		if arg == "version" {
			printVersion()
			return nil
		}
	}

	zap.L().Info("calendar service is running...")
	cfg, err := config.NewConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error during config read: %w", err)
	}

	err = logger.InitLogger(cfg.Logger)
	if err != nil {
		return fmt.Errorf("erro during logger init: %w", err)
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	notifyCtx, stop := signal.NotifyContext(cancelCtx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	var repo api.EventRepository
	if cfg.Storage.UseMemoryStorage {
		repo = memorystorage.New()
	} else {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s database=%s password=%s sslmode=disable",
			cfg.Storage.DB.Host,
			cfg.Storage.DB.Port,
			cfg.Storage.DB.Username,
			cfg.Storage.DB.DB,
			cfg.Storage.DB.Password)
		dbStorage := sqlstorage.New()
		if err := dbStorage.Connect(notifyCtx, dsn); err != nil {
			return fmt.Errorf("failed to init db storage: %w", err)
		}
		repo = dbStorage
		defer func() {
			if err := dbStorage.Close(); err != nil {
				zap.L().Error("error during closing db storage", zap.Error(err))
			}
		}()
	}
	zap.L().Info("calendar service storage started...")

	apiService := api.New(cfg, repo)
	server := internalhttp.NewServer(cfg, apiService)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go gracefulShutdown(notifyCtx, server, &wg)

	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.L().Error("Failed to start http server", zap.Error(err))
		// manually calling server shutdown
		cancelFunc()
	}

	// checking all server connections surely canceled
	wg.Wait()
	zap.L().Info("calendar service stopped")
	return nil
}

func gracefulShutdown(notifyCtx context.Context, server *internalhttp.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	<-notifyCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		// Error from closing listeners, or context timeout
		zap.L().Error("Failed to stop http server", zap.Error(err))
	}
}
