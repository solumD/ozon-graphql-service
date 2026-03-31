package graphql

import (
	"context"

	"github.com/solumD/ozon-grapql-service/internal/model"
)

type PostUsecase interface {
	CreatePost(ctx context.Context, userUUID string, title string, content string, commentsEnabled bool) (model.Post, error)
	GetPost(ctx context.Context, id int64) (model.Post, error)
	ListPosts(ctx context.Context) ([]model.Post, error)
	ChangeCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error)
}

type CommentUsecase interface {
	CreateComment(ctx context.Context, userUUID string, postID int64, parentID *int64, content string) (model.Comment, error)
	ListComments(ctx context.Context, filter model.CommentListFilter) (model.CommentConnection, error)
}
