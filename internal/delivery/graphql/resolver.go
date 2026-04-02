package graphql

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/solumD/ozon-grapql-service/internal/delivery/graphql/generated"
)

const kapInterval = 10 * time.Second

type Resolver struct {
	postUsecase     PostUsecase
	commentUsecase  CommentUsecase
	commentConsumer CommentConsumer
	log             *slog.Logger
}

func NewResolver(postUsecase PostUsecase, commentUsecase CommentUsecase, commentConsumer CommentConsumer, log *slog.Logger) *Resolver {
	return &Resolver{
		postUsecase:     postUsecase,
		commentUsecase:  commentUsecase,
		commentConsumer: commentConsumer,
		log:             log,
	}
}

// RegisterRoutes регистрирует роуты для GraphQL
func RegisterRoutes(r chi.Router, resolver *Resolver, playgroundEnabled bool) {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},

		KeepAlivePingInterval: kapInterval,
	})

	r.Handle("/graphql", srv)
	if playgroundEnabled {
		r.Get("/playground", func(w http.ResponseWriter, req *http.Request) {
			playground.Handler("GraphQL playground", "/graphql").ServeHTTP(w, req)
		})
	}
}
