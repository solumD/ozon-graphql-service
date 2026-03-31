package memory

import (
	"context"
	"sort"
	"time"

	"github.com/solumD/ozon-grapql-service/internal/model"
)

type PostRepository struct {
	storage *Storage
}

func NewPostRepository(storage *Storage) *PostRepository {
	return &PostRepository{storage: storage}
}

func (r *PostRepository) Create(_ context.Context, post model.Post) (model.Post, error) {
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

	return created, nil
}

func (r *PostRepository) GetByID(_ context.Context, id int64) (model.Post, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	post, ok := r.storage.posts[id]
	if !ok {
		return model.Post{}, model.ErrPostNotFound
	}

	return post, nil
}

func (r *PostRepository) List(_ context.Context) ([]model.Post, error) {
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

	return posts, nil
}

func (r *PostRepository) UpdateCommentsAvailability(_ context.Context, postID int64, enabled bool) (model.Post, error) {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	post, ok := r.storage.posts[postID]
	if !ok {
		return model.Post{}, model.ErrPostNotFound
	}

	post.CommentsEnabled = enabled

	r.storage.posts[postID] = post

	return post, nil
}
