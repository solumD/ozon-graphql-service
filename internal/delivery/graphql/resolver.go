package graphql

import (
	"log/slog"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/solumD/ozon-grapql-service/internal/delivery/graphql/generated"
)

type Resolver struct {
	postUsecase    PostUsecase
	commentUsecase CommentUsecase
	log            *slog.Logger
}

func NewResolver(postUsecase PostUsecase, commentUsecase CommentUsecase, log *slog.Logger) *Resolver {
	return &Resolver{
		postUsecase:    postUsecase,
		commentUsecase: commentUsecase,
		log:            log,
	}
}

func RegisterRoutes(r chi.Router, resolver *Resolver, playgroundEnabled bool) {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	r.Handle("/graphql", srv)
	if playgroundEnabled {
		r.Get("/playground", func(w http.ResponseWriter, req *http.Request) {
			playground.Handler("GraphQL playground", "/graphql").ServeHTTP(w, req)
		})
	}
}
