package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
)

type CommentUsecase struct {
	postRepository    PostRepository
	commentRepository CommentRepository
}

func NewCommentUsecase(postRepository PostRepository, commentRepository CommentRepository) *CommentUsecase {
	return &CommentUsecase{
		postRepository:    postRepository,
		commentRepository: commentRepository,
	}
}

func (uc *CommentUsecase) CreateComment(ctx context.Context, userUUID string, postID int64, parentID *int64, content string) (model.Comment, error) {
	content = strings.TrimSpace(content)
	userUUID = strings.TrimSpace(userUUID)

	if content == "" {
		return model.Comment{}, coreerrors.ErrEmptyCommentContent
	}

	if len(content) > model.MaxCommentLength {
		return model.Comment{}, coreerrors.ErrCommentTooLong
	}

	post, err := uc.postRepository.GetByID(ctx, postID)
	if err != nil {
		return model.Comment{}, err
	}

	if !post.CommentsEnabled {
		return model.Comment{}, coreerrors.ErrCommentsDisabled
	}

	if parentID != nil {
		parentComment, err := uc.commentRepository.GetByID(ctx, *parentID)
		if err != nil {
			return model.Comment{}, err
		}

		if parentComment.PostID != postID {
			return model.Comment{}, coreerrors.ErrInvalidParentComment
		}
	}

	comment := model.Comment{
		UserUUID: userUUID,
		PostID:   postID,
		ParentID: utils.CloneInt64Ptr(parentID),
		Content:  content,
	}

	return uc.commentRepository.Create(ctx, comment)
}

func (uc *CommentUsecase) ListComments(ctx context.Context, filter model.CommentListFilter) (model.CommentConnection, error) {
	normalizedFilter, err := normalizeCommentListFilter(filter)
	if err != nil {
		return model.CommentConnection{}, err
	}

	comments, hasNextPage, err := uc.commentRepository.ListByPostAndParent(ctx, normalizedFilter)
	if err != nil {
		return model.CommentConnection{}, err
	}

	edges := make([]model.CommentEdge, 0, len(comments))
	for _, comment := range comments {
		cursor := encodeCursor(model.Cursor{CreatedAt: comment.CreatedAt, ID: comment.ID})
		edges = append(edges, model.CommentEdge{
			Cursor: cursor,
			Node:   comment,
		})
	}

	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	return model.CommentConnection{
		Edges: edges,
		PageInfo: model.PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   endCursor,
		},
	}, nil
}

func normalizeCommentListFilter(filter model.CommentListFilter) (model.CommentListFilter, error) {
	if filter.PostID <= 0 {
		return model.CommentListFilter{}, coreerrors.ErrInvalidPagination
	}

	if filter.First <= 0 {
		filter.First = model.DefaultCommentsPageSize
	}

	if filter.First > model.MaxCommentsPageSize {
		filter.First = model.MaxCommentsPageSize
	}

	return filter, nil
}

func encodeCursor(cursor model.Cursor) string {
	raw := fmt.Sprintf("%d:%d", cursor.CreatedAt.UnixNano(), cursor.ID)
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))

	return encoded
}

func DecodeCursor(raw string) (model.Cursor, error) {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return model.Cursor{}, coreerrors.ErrInvalidCursor
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return model.Cursor{}, coreerrors.ErrInvalidCursor
	}

	unixNano, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, coreerrors.ErrInvalidCursor
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return model.Cursor{}, coreerrors.ErrInvalidCursor
	}

	return model.Cursor{
		CreatedAt: time.Unix(0, unixNano).UTC(),
		ID:        id,
	}, nil
}
