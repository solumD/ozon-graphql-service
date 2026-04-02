package usecase

import (
	"context"

	"github.com/solumD/ozon-graphql-service/internal/model"
)

type PostRepository interface {
	Create(ctx context.Context, post model.Post) (model.Post, error)
	GetByID(ctx context.Context, id int64) (model.Post, error)
	List(ctx context.Context) ([]model.Post, error)
	UpdateCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment model.Comment) (model.Comment, error)
	GetByID(ctx context.Context, id int64) (model.Comment, error)
	ListByPostAndParent(ctx context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error)
}

type CommentProducer interface {
	PublishComment(ctx context.Context, comment model.Comment)
}
