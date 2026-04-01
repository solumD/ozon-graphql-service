package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
	pg "github.com/solumD/ozon-grapql-service/pkg/postgres"
)

type commentRepository struct {
	db *pg.Postgres
}

func NewCommentRepository(db *pg.Postgres) *commentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment model.Comment) (model.Comment, error) {
	query := `
		INSERT INTO comments (user_uuid, post_id, parent_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.Pool().QueryRow(ctx, query, comment.UserUUID, comment.PostID, comment.ParentID, comment.Content).
		Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return model.Comment{}, coreerrors.ErrCreateComment
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
	comment.HasReplies = false

	return comment, nil
}

func (r *commentRepository) GetByID(ctx context.Context, id int64) (model.Comment, error) {
	query := `
		SELECT c.id, c.user_uuid, c.post_id, c.parent_id, c.content, c.created_at,
		       EXISTS (SELECT 1 FROM comments child WHERE child.parent_id = c.id) AS has_replies
		FROM comments c
		WHERE c.id = $1
	`

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
			return model.Comment{}, coreerrors.ErrCommentNotFound
		}

		return model.Comment{}, coreerrors.ErrGetComment
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)

	return comment, nil
}

func (r *commentRepository) ListByPostAndParent(ctx context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
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

	rows, err := r.db.Pool().Query(ctx, query, args...)
	if err != nil {
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
			return nil, false, coreerrors.ErrListComments
		}

		comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, false, coreerrors.ErrListComments
	}

	hasNextPage := len(comments) > limit
	if hasNextPage {
		comments = comments[:limit]
	}

	return comments, hasNextPage, nil
}
