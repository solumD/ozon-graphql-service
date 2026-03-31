package usecase

import (
	"context"
	"strings"

	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
)

type PostUsecase struct {
	postRepository PostRepository
}

func NewPostUsecase(postRepository PostRepository) *PostUsecase {
	return &PostUsecase{postRepository: postRepository}
}

func (uc *PostUsecase) CreatePost(ctx context.Context, userUUID string, title string, content string, commentsEnabled bool) (model.Post, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	userUUID = strings.TrimSpace(userUUID)

	if title == "" {
		return model.Post{}, coreerrors.ErrEmptyPostTitle
	}

	if content == "" {
		return model.Post{}, coreerrors.ErrEmptyPostContent
	}

	post := model.Post{
		UserUUID:        userUUID,
		Title:           title,
		Content:         content,
		CommentsEnabled: commentsEnabled,
	}

	return uc.postRepository.Create(ctx, post)
}

func (uc *PostUsecase) GetPost(ctx context.Context, id int64) (model.Post, error) {
	return uc.postRepository.GetByID(ctx, id)
}

func (uc *PostUsecase) ListPosts(ctx context.Context) ([]model.Post, error) {
	return uc.postRepository.List(ctx)
}

func (uc *PostUsecase) ChangeCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error) {
	return uc.postRepository.UpdateCommentsAvailability(ctx, postID, enabled)
}
