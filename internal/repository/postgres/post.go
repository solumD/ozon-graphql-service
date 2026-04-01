package postgres

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
	pg "github.com/solumD/ozon-grapql-service/pkg/postgres"
)

type postRepository struct {
	db  *pg.Postgres
	log *slog.Logger
}

func NewPostRepository(db *pg.Postgres, log *slog.Logger) *postRepository {
	return &postRepository{db: db, log: log}
}

func (r *postRepository) Create(ctx context.Context, post model.Post) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		INSERT INTO posts (user_uuid, title, content, comments_enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	log.Debug("exuting query", logger.String("query", query))

	err := r.db.Pool().QueryRow(ctx, query, post.UserUUID, post.Title, post.Content, post.CommentsEnabled).
		Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		log.Error("failed to create post", logger.Error(err))
		return model.Post{}, coreerrors.ErrCreatePost
	}

	log.Info("post created", logger.Int64("post_id", post.ID), logger.String("user_uuid", post.UserUUID))

	return post, nil
}

func (r *postRepository) GetByID(ctx context.Context, id int64) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		SELECT id, user_uuid, title, content, comments_enabled, created_at
		FROM posts
		WHERE id = $1
	`

	log.Debug("exuting query", logger.String("query", query))

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
			log.Warn("post not found", logger.Int64("post_id", id))
			return model.Post{}, coreerrors.ErrPostNotFound
		}

		log.Error("failed to get post", logger.Error(err), logger.Int64("post_id", id))
		return model.Post{}, coreerrors.ErrGetPost
	}

	log.Info("post retrieved", logger.Int64("post_id", id))

	return post, nil
}

func (r *postRepository) List(ctx context.Context) ([]model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		SELECT id, user_uuid, title, content, comments_enabled, created_at
		FROM posts
		ORDER BY created_at ASC, id ASC
	`

	log.Debug("exuting query", logger.String("query", query))

	rows, err := r.db.Pool().Query(ctx, query)
	if err != nil {
		log.Error("failed to list posts", logger.Error(err))

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
			log.Error("failed to scan post row", logger.Error(err))
			return nil, err
		}

		posts = append(posts, post)
	}

	err = rows.Err()
	if err != nil {
		log.Error("failed while iterating posts", logger.Error(err))

		return []model.Post{}, coreerrors.ErrListPosts
	}

	log.Info("posts listed", logger.Int64("count", int64(len(posts))))

	return posts, nil
}

func (r *postRepository) UpdateCommentsAvailability(ctx context.Context, postID int64, enabled bool) (model.Post, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		UPDATE posts
		SET comments_enabled = $2
		WHERE id = $1
		RETURNING id, user_uuid, title, content, comments_enabled, created_at
	`

	log.Debug("exuting query", logger.String("query", query))

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
			log.Warn("post not found for comments availability update", logger.Int64("post_id", postID))

			return model.Post{}, coreerrors.ErrPostNotFound
		}

		log.Error("failed to update comments availability", logger.Error(err), logger.Int64("post_id", postID))

		return model.Post{}, coreerrors.ErrUpdateCommentsAvailability
	}

	log.Info("comments availability updated", logger.Int64("post_id", postID), logger.Any("enabled", enabled))

	return post, nil
}
