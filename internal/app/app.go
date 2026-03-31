package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solumD/ozon-grapql-service/config"
	graphql "github.com/solumD/ozon-grapql-service/internal/delivery/graphql"
	"github.com/solumD/ozon-grapql-service/internal/delivery/router"
	"github.com/solumD/ozon-grapql-service/internal/repository/memory"
	"github.com/solumD/ozon-grapql-service/internal/usecase"
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

	storage := memory.NewStorage()
	postRepository := memory.NewPostRepository(storage)
	commentRepository := memory.NewCommentRepository(storage)

	postUsecase := usecase.NewPostUsecase(postRepository)
	commentUsecase := usecase.NewCommentUsecase(postRepository, commentRepository)

	resolver := graphql.NewResolver(postUsecase, commentUsecase)

	router := router.NewRouter()
	graphql.RegisterRoutes(router, resolver, cfg.PlaygroundEnabled)

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
