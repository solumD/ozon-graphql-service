package memory

import (
	"context"
	"log/slog"
	"sort"
	"time"

	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
)

type PostRepository struct {
	storage *Storage
	log     *slog.Logger
}

func NewPostRepository(storage *Storage, log *slog.Logger) *PostRepository {
	return &PostRepository{storage: storage, log: log}
}

func (r *PostRepository) Create(_ context.Context, post model.Post) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	r.storage.nextPostID++

	created := model.Post{
		ID:              r.storage.nextPostID,
		UserUUID:        post.UserUUID,
		Title:           post.Title,
		Content:         post.Content,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       time.Now().UTC(),
	}

	r.storage.posts[created.ID] = created

	log.Info("post created", logger.Int64("post_id", created.ID), logger.String("user_uuid", created.UserUUID))

	return created, nil
}

func (r *PostRepository) GetByID(_ context.Context, id int64) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	post, ok := r.storage.posts[id]
	if !ok {
		log.Warn("post not found", logger.Int64("post_id", id))

		return model.Post{}, coreerrors.ErrPostNotFound
	}

	log.Info("post retrieved", logger.Int64("post_id", id))

	return post, nil
}

func (r *PostRepository) List(_ context.Context) ([]model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	posts := make([]model.Post, 0, len(r.storage.posts))
	for _, post := range r.storage.posts {
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		if posts[i].CreatedAt.Equal(posts[j].CreatedAt) {
			return posts[i].ID < posts[j].ID
		}

		return posts[i].CreatedAt.Before(posts[j].CreatedAt)
	})

	log.Info("posts listed", logger.Int64("count", int64(len(posts))))

	return posts, nil
}

func (r *PostRepository) UpdateCommentsAvailability(_ context.Context, postID int64, enabled bool) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	post, ok := r.storage.posts[postID]
	if !ok {
		log.Warn("post not found for comments availability update", logger.Int64("post_id", postID))

		return model.Post{}, coreerrors.ErrPostNotFound
	}

	post.CommentsEnabled = enabled
	r.storage.posts[postID] = post

	log.Info("comments availability updated", logger.Int64("post_id", postID), logger.Any("enabled", enabled))

	return post, nil
}
