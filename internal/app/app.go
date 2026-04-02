package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/solumD/ozon-grapql-service/config"
	inmemoryBroker "github.com/solumD/ozon-grapql-service/internal/broker/in_memory"
	graphql "github.com/solumD/ozon-grapql-service/internal/delivery/graphql"
	"github.com/solumD/ozon-grapql-service/internal/delivery/router"
	inmemoryRepo "github.com/solumD/ozon-grapql-service/internal/repository/in_memory"
	pgrepo "github.com/solumD/ozon-grapql-service/internal/repository/postgres"
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

	lg := logger.NewLogger(cfg.LoggerLevel)
	lg.Debug("debug messages are enabled")

	lg.Info("configuration loaded", logger.Any("storage_type", cfg.StorageType))

	var (
		postRepository    usecase.PostRepository
		commentRepository usecase.CommentRepository
	)

	commentBroker := inmemoryBroker.NewCommentBroker()

	// настройка хранилища
	if cfg.StorageType == config.StorageTypePostgres {
		postgresConn := pg.New(cfg.PostgresDSN)
		if err := postgresConn.Ping(ctx); err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer postgresConn.Close()

		lg.Info("connected to postgres")

		postRepository = pgrepo.NewPostRepository(postgresConn, lg)
		commentRepository = pgrepo.NewCommentRepository(postgresConn, lg)
	} else {
		storage := inmemoryRepo.NewStorage()
		postRepository = inmemoryRepo.NewPostRepository(storage, lg)
		commentRepository = inmemoryRepo.NewCommentRepository(storage, lg)
	}

	postUsecase := usecase.NewPostUsecase(postRepository, lg)
	commentUsecase := usecase.NewCommentUsecase(postRepository, commentRepository, commentBroker, lg)
	resolver := graphql.NewResolver(postUsecase, commentUsecase, commentBroker, lg)

	router := router.NewRouter()

	graphql.RegisterRoutes(router, resolver, cfg.PlaygroundEnabled)

	server := httpserver.New(cfg.ServerAddr(), router)
	server.Run()

	lg.Info("server started", logger.String("addr", cfg.ServerAddr()))

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	<-interrupt

	lg.Info("shutting down server")

	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdownCtx()

	if err := server.Shutdown(shutdownCtx); err != nil {
		lg.Error("error while shutting down server", logger.Error(err))
	}

	lg.Info("server stopped")
}
