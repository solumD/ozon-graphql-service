package usecase

import (
	"context"
	"log/slog"
	"strings"

	coreerrors "github.com/solumD/ozon-graphql-service/internal/core_errors"
	"github.com/solumD/ozon-graphql-service/internal/model"
	"github.com/solumD/ozon-graphql-service/internal/utils"
	"github.com/solumD/ozon-graphql-service/pkg/logger"
)

type PostUsecase struct {
	postRepository PostRepository
	log            *slog.Logger
}

func NewPostUsecase(postRepository PostRepository, log *slog.Logger) *PostUsecase {
	return &PostUsecase{postRepository: postRepository, log: log}
}

// CreatePost создает новый пост
func (uc *PostUsecase) CreatePost(ctx context.Context, userUUID string, title string, content string, commentsEnabled bool) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	userUUID = strings.TrimSpace(userUUID)

	if title == "" {
		log.Warn("empty post title")

		return model.Post{}, coreerrors.ErrEmptyPostTitle
	}

	if content == "" {
		log.Warn("empty post content")

		return model.Post{}, coreerrors.ErrEmptyPostContent
	}

	post := model.Post{
		UserUUID:        userUUID,
		Title:           title,
		Content:         content,
		CommentsEnabled: commentsEnabled,
	}

	created, err := uc.postRepository.Create(ctx, post)
	if err != nil {
		log.Error("failed to create post", logger.Error(err))

		return model.Post{}, err
	}

	log.Info("post created in usecase", logger.Int64("post_id", created.ID))

	return created, nil
}

// GetPost возвращает пост по id
func (uc *PostUsecase) GetPost(ctx context.Context, id int64) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	post, err := uc.postRepository.GetByID(ctx, id)
	if err != nil {
		log.Error("failed to get post", logger.Error(err), logger.Int64("post_id", id))

		return model.Post{}, err
	}

	log.Info("post returned from usecase", logger.Int64("post_id", id))

	return post, nil
}

// ListPosts возвращает список всех постов
func (uc *PostUsecase) ListPosts(ctx context.Context) ([]model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	posts, err := uc.postRepository.List(ctx)
	if err != nil {
		log.Error("failed to list posts", logger.Error(err))

		return nil, err
	}

	log.Info("posts returned from usecase", logger.Int64("count", int64(len(posts))))

	return posts, nil
}

// ChangeCommentsAvailability изменяет доступность комментариев у
func (uc *PostUsecase) ChangeCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	post, err := uc.postRepository.UpdateCommentsAvailability(ctx, postID, enabled)
	if err != nil {
		log.Error("failed to change comments availability", logger.Error(err), logger.Int64("post_id", postID))

		return model.Post{}, err
	}

	log.Info("comments availability changed in usecase", logger.Int64("post_id", postID), logger.Any("enabled", enabled))

	return post, nil
}
