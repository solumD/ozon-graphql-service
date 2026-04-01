package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
)

type CommentUsecase struct {
	postRepository    PostRepository
	commentRepository CommentRepository
	log               *slog.Logger
}

func NewCommentUsecase(postRepository PostRepository, commentRepository CommentRepository, log *slog.Logger) *CommentUsecase {
	return &CommentUsecase{
		postRepository:    postRepository,
		commentRepository: commentRepository,
		log:               log,
	}
}

func (uc *CommentUsecase) CreateComment(ctx context.Context, userUUID string, postID int64, parentID *int64, content string) (model.Comment, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	content = strings.TrimSpace(content)
	userUUID = strings.TrimSpace(userUUID)

	if content == "" {
		log.Warn("empty comment content")

		return model.Comment{}, coreerrors.ErrEmptyCommentContent
	}

	if len(content) > model.MaxCommentLength {
		log.Warn("comment too long", logger.Int64("length", int64(len(content))))

		return model.Comment{}, coreerrors.ErrCommentTooLong
	}

	post, err := uc.postRepository.GetByID(ctx, postID)
	if err != nil {
		log.Error("failed to get post for comment creation", logger.Error(err), logger.Int64("post_id", postID))

		return model.Comment{}, err
	}

	if !post.CommentsEnabled {
		log.Warn("comments are disabled", logger.Int64("post_id", postID))

		return model.Comment{}, coreerrors.ErrCommentsDisabled
	}

	if parentID != nil {
		parentComment, err := uc.commentRepository.GetByID(ctx, *parentID)
		if err != nil {
			log.Error("failed to get parent comment", logger.Error(err), logger.Int64("parent_id", *parentID))

			return model.Comment{}, err
		}

		if parentComment.PostID != postID {
			log.Warn("invalid parent comment", logger.Int64("parent_id", *parentID), logger.Int64("post_id", postID))

			return model.Comment{}, coreerrors.ErrInvalidParentComment
		}
	}

	comment := model.Comment{
		UserUUID: userUUID,
		PostID:   postID,
		ParentID: utils.CloneInt64Ptr(parentID),
		Content:  content,
	}

	created, err := uc.commentRepository.Create(ctx, comment)
	if err != nil {
		log.Error("failed to create comment", logger.Error(err), logger.Int64("post_id", postID))

		return model.Comment{}, err
	}

	log.Info("comment created in usecase", logger.Int64("comment_id", created.ID), logger.Int64("post_id", created.PostID))

	return created, nil
}

func (uc *CommentUsecase) ListComments(ctx context.Context, filter model.CommentListFilter) (model.CommentConnection, error) {
	fn := utils.GetCurrentFunctionName()
	log := uc.log.With(logger.String("fn", fn))

	normalizedFilter, err := normalizeCommentListFilter(filter)
	if err != nil {
		log.Error("failed to normalize comment filter", logger.Error(err))

		return model.CommentConnection{}, err
	}

	comments, hasNextPage, err := uc.commentRepository.ListByPostAndParent(ctx, normalizedFilter)
	if err != nil {
		log.Error("failed to list comments", logger.Error(err), logger.Int64("post_id", normalizedFilter.PostID))

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

	log.Info("comments returned from usecase", logger.Int64("post_id", normalizedFilter.PostID), logger.Int64("count", int64(len(edges))), logger.Any("has_next_page", hasNextPage))

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
