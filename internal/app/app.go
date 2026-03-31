package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solumD/ozon-grapql-service/config"
	deliveryhttp "github.com/solumD/ozon-grapql-service/internal/delivery/http"
	httpserver "github.com/solumD/ozon-grapql-service/pkg/http_server"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
	pg "github.com/solumD/ozon-grapql-service/pkg/postgres"
)

const shutdownTimeout = 10 * time.Second

func InitAndRun(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.MustLoad()

	logg := logger.NewLogger(cfg.LoggerLevel)
	logg.Info("configuration loaded", logger.Any("storage_type", cfg.StorageType))

	if cfg.StorageType == config.StorageTypePostgres {
		postgresConn := pg.New(cfg.PostgresDSN)
		if err := postgresConn.Ping(ctx); err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer postgresConn.Close()

		logg.Info("connected to postgres")
	}

	router := deliveryhttp.NewRouter(ctx)
	server := httpserver.New(cfg.ServerAddr(), router)
	server.Run()

	logg.Info("server started", logger.String("addr", cfg.ServerAddr()))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	logg.Info("shutting down server")

	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdownCtx()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logg.Error("error while shutting down server", logger.Error(err))
	}

	logg.Info("server stopped")
}
