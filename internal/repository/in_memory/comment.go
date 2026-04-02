package inmemory

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

type CommentRepository struct {
	storage *Storage
	log     *slog.Logger
}

func NewCommentRepository(storage *Storage, log *slog.Logger) *CommentRepository {
	return &CommentRepository{storage: storage, log: log}
}

// Create сохраняет комментарии в хранилище
func (r *CommentRepository) Create(_ context.Context, comment model.Comment) (model.Comment, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	r.storage.nextCommentID++

	created := model.Comment{
		ID:         r.storage.nextCommentID,
		UserUUID:   comment.UserUUID,
		PostID:     comment.PostID,
		ParentID:   utils.CloneInt64Ptr(comment.ParentID),
		HasReplies: false,
		Content:    comment.Content,
		CreatedAt:  time.Now().UTC(),
	}

	r.storage.comments[created.ID] = created

	// создаем отдельный бакет для ответов на комментарий
	key := r.storage.makeCommentBucketKey(created.PostID, created.ParentID)
	r.storage.commentsByBucket[key] = append(r.storage.commentsByBucket[key], created.ID)

	log.Info("comment created", logger.Int64("comment_id", created.ID), logger.Int64("post_id", created.PostID))

	return created, nil
}

// GetByID возвращает комментарий по его id из хранилища
func (r *CommentRepository) GetByID(_ context.Context, id int64) (model.Comment, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	comment, ok := r.storage.comments[id]
	if !ok {
		log.Warn("comment not found", logger.Int64("comment_id", id))

		return model.Comment{}, coreerrors.ErrCommentNotFound
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
	comment.HasReplies = r.hasReplies(comment.PostID, comment.ID)

	log.Info("comment retrieved", logger.Int64("comment_id", id))

	return comment, nil
}

// ListByPostAndParent возвращает список комментариев по post_id и parent_id
func (r *CommentRepository) ListByPostAndParent(_ context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
	fn := utils.GetCurrentFunctionName()
	log := r.log.With(logger.String("fn", fn))

	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	// получаем бакет с ответами на комментарий
	key := r.storage.makeCommentBucketKey(filter.PostID, filter.ParentID)
	bucket := r.storage.commentsByBucket[key]

	comments := make([]model.Comment, 0, len(bucket))
	for _, commentID := range bucket {
		comment, ok := r.storage.comments[commentID]
		if !ok {
			continue
		}

		comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)
		comment.HasReplies = r.hasReplies(comment.PostID, comment.ID)

		comments = append(comments, comment)
	}

	sort.Slice(comments, func(i, j int) bool {
		if comments[i].CreatedAt.Equal(comments[j].CreatedAt) {
			return comments[i].ID < comments[j].ID
		}

		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	startIdx := 0
	if filter.After != nil {
		for i, comment := range comments {
			if comment.CreatedAt.After(filter.After.CreatedAt) ||
				(comment.CreatedAt.Equal(filter.After.CreatedAt) && comment.ID > filter.After.ID) {

				startIdx = i
				break
			}

			startIdx = len(comments)
		}
	}

	if startIdx > len(comments) {
		startIdx = len(comments)
	}

	first := filter.First
	if first <= 0 {
		first = model.DefaultCommentsPageSize
	}

	endIdx := startIdx + first
	hasNextPage := endIdx < len(comments)
	if endIdx > len(comments) {
		endIdx = len(comments)
	}

	log.Info("comments listed", logger.Int64("post_id", filter.PostID), logger.Int64("count", int64(endIdx-startIdx)), logger.Any("has_next_page", hasNextPage))

	return comments[startIdx:endIdx], hasNextPage, nil
}

// hasReplies возвращает true, если у комментария есть дочерние комментарии
func (r *CommentRepository) hasReplies(postID, commentID int64) bool {
	key := r.storage.makeCommentBucketKey(postID, &commentID)
	children := r.storage.commentsByBucket[key]

	return len(children) > 0
}
