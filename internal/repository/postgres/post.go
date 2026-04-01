package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	pg "github.com/solumD/ozon-grapql-service/pkg/postgres"
)

type postRepository struct {
	db *pg.Postgres
}

func NewPostRepository(db *pg.Postgres) *postRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post model.Post) (model.Post, error) {
	query := `
		INSERT INTO posts (user_uuid, title, content, comments_enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.Pool().QueryRow(ctx, query, post.UserUUID, post.Title, post.Content, post.CommentsEnabled).
		Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		return model.Post{}, coreerrors.ErrCreatePost
	}

	return post, nil
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (model.Post, error) {
	query := `
		SELECT id, user_uuid, title, content, comments_enabled, created_at
		FROM posts
		WHERE id = $1
	`

	var post model.Post
	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.UserUUID,
		&post.Title,
		&post.Content,
		&post.CommentsEnabled,
		&post.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Post{}, coreerrors.ErrPostNotFound
		}

		return model.Post{}, coreerrors.ErrGetPost
	}

	return post, nil
}

func (r *postRepository) List(ctx context.Context) ([]model.Post, error) {
	query := `
		SELECT id, user_uuid, title, content, comments_enabled, created_at
		FROM posts
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		return []model.Post{}, coreerrors.ErrListPosts
	}
	defer rows.Close()

	posts := make([]model.Post, 0)
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.UserUUID,
			&post.Title,
			&post.Content,
			&post.CommentsEnabled,
			&post.CreatedAt,
		); err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		return []model.Post{}, coreerrors.ErrListPosts
	}

	return posts, nil
}

func (r *postRepository) UpdateCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error) {
	query := `
		UPDATE posts
		SET comments_enabled = $2
		WHERE id = $1
		RETURNING id, user_uuid, title, content, comments_enabled, created_at
	`

	var post model.Post
	err := r.db.Pool().QueryRow(ctx, query, postID, enabled).Scan(
		&post.ID,
		&post.UserUUID,
		&post.Title,
		&post.Content,
		&post.CommentsEnabled,
		&post.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Post{}, coreerrors.ErrPostNotFound
		}

		return model.Post{}, coreerrors.ErrUpdateCommentsAvailability
	}

	return post, nil
}
