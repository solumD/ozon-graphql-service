package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	coreerrors "github.com/solumD/ozon-graphql-service/internal/core_errors"
	"github.com/solumD/ozon-graphql-service/internal/model"
	"github.com/solumD/ozon-graphql-service/internal/utils"
	"github.com/solumD/ozon-graphql-service/pkg/logger"
	pg "github.com/solumD/ozon-graphql-service/pkg/postgres"
)

type commentRepository struct {
	db  *pg.Postgres
	log *slog.Logger
}

func NewCommentRepository(db *pg.Postgres, log *slog.Logger) *commentRepository {
	return &commentRepository{db: db, log: log}
}

func (r *commentRepository) Create(ctx context.Context, comment model.Comment) (model.Comment, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		INSERT INTO comments (user_uuid, post_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	log.Debug("executing query", logger.String("query", query))

	err := r.db.Pool().QueryRow(ctx, query, comment.UserUUID, comment.PostID, comment.ParentID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		log.Error("failed to create comment", logger.Error(err))

		return model.Comment{}, coreerrors.ErrCreateComment
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
	comment.HasReplies = false

	log.Info("comment created", logger.Int64("comment_id", comment.ID), logger.Int64("post_id", comment.PostID))

	return comment, nil
}

func (r *commentRepository) GetByID(ctx context.Context, id int64) (model.Comment, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	query := `
		SELECT c.id, c.user_uuid, c.post_id, c.parent_id, c.content, c.created_at,
		       EXISTS (SELECT 1 FROM comments child WHERE child.parent_id = c.id) AS has_replies
		FROM comments c
		WHERE c.id = $1
	`

	log.Debug("executing query", logger.String("query", query))

	var comment model.Comment
	err := r.db.Pool().QueryRow(ctx, query, id).Scan(
		&comment.ID,
		&comment.UserUUID,
		&comment.PostID,
		&comment.ParentID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.HasReplies,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("comment not found", logger.Int64("comment_id", id))

			return model.Comment{}, coreerrors.ErrCommentNotFound
		}

		log.Error("failed to get comment", logger.Error(err), logger.Int64("comment_id", id))

		return model.Comment{}, coreerrors.ErrGetComment
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)

	log.Info("comment retrieved", logger.Int64("comment_id", id))

	return comment, nil
}

func (r *commentRepository) ListByPostAndParent(ctx context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	args := []any{filter.PostID}
	whereParent := "c.parent_id IS NULL"
	if filter.ParentID != nil {
		whereParent = fmt.Sprintf("c.parent_id = $%d", len(args)+1)
		args = append(args, *filter.ParentID)
	}

	cursorClause := ""
	if filter.After != nil {
		cursorClause = fmt.Sprintf(" AND (c.created_at, c.id) > ($%d, $%d)", len(args)+1, len(args)+2)
		args = append(args, filter.After.CreatedAt, filter.After.ID)
	}

	limit := filter.First
	if limit <= 0 {
		limit = model.DefaultCommentsPageSize
	}
	args = append(args, limit+1)

	query := fmt.Sprintf(`
		SELECT c.id, c.user_uuid, c.post_id, c.parent_id, c.content, c.created_at,
		       EXISTS (SELECT 1 FROM comments child WHERE child.parent_id = c.id) AS has_replies
		FROM comments c
		WHERE c.post_id = $1 AND %s%s
		ORDER BY c.created_at ASC, c.id ASC
		LIMIT $%d
	`, whereParent, cursorClause, len(args))

	log.Debug("executing query", logger.String("query", query))

	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
		log.Error("failed to list comments", logger.Error(err), logger.Int64("post_id", filter.PostID))

		return nil, false, coreerrors.ErrListComments
	}
	defer rows.Close()

	comments := make([]model.Comment, 0)
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.UserUUID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.HasReplies,
		); err != nil {
			log.Error("failed to scan comment row", logger.Error(err))

			return nil, false, coreerrors.ErrListComments
		}

		comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Error("failed while iterating comments", logger.Error(err))

		return nil, false, coreerrors.ErrListComments
	}

	hasNextPage := len(comments) > limit
	if hasNextPage {
		comments = comments[:limit]
	}

	log.Info("comments listed", logger.Int64("post_id", filter.PostID), logger.Int64("count", int64(len(comments))), logger.Any("has_next_page", hasNextPage))

	return comments, hasNextPage, nil
}
