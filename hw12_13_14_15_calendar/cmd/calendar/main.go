package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/app"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/logger"
	"github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const ServerShutdownTimeout = time.Second * 3

var configFilePath string

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

	notifyCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	var repo app.EventRepository
	if cfg.Storage.UseMemoryStorage {
		repo = memorystorage.NewMemStorage()
	} else {
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s database=%s password=%s sslmode=disable",
			cfg.Storage.DB.Host,
			cfg.Storage.DB.Port,
			cfg.Storage.DB.Username,
			cfg.Storage.DB.DB,
			cfg.Storage.DB.Password)
		dbStorage := sqlstorage.NewDBStorage()
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

	apiService := app.New(repo)
	httpAPI := internalhttp.NewHTTPApi(cfg.API.HTTP, apiService)
	grpcAPI := grpc.NewGRPCApi(cfg.API.GRPC, apiService)

	wg := sync.WaitGroup{}
	wg.Add(4)

	go shutdownHTTP(notifyCtx, httpAPI, &wg)
	go func() {
		defer wg.Done()
		httpAPI.Start(stop)
	}()

	go shutdownGRPC(notifyCtx, grpcAPI, &wg)
	go func() {
		defer wg.Done()
		grpcAPI.Start(stop)
	}()

	// checking all server connections surely canceled
	wg.Wait()
	zap.L().Info("calendar service stopped")
	return nil
}

func shutdownHTTP(ctx context.Context, api *internalhttp.API, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()

	if err := api.Stop(ctx); err != nil {
		// Error from closing listeners, or context timeout
		zap.L().Error("Error during stopping http server", zap.Error(err))
	}
}

func shutdownGRPC(ctx context.Context, api *grpc.API, wg *sync.WaitGroup) {
	defer wg.Done()
	<-ctx.Done()
	api.Stop()
}
