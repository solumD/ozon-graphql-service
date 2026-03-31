package memory

import (
	"context"
	"sort"
	"time"

	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
)

type CommentRepository struct {
	storage *Storage
}

func NewCommentRepository(storage *Storage) *CommentRepository {
	return &CommentRepository{storage: storage}
}

func (r *CommentRepository) Create(_ context.Context, comment model.Comment) (model.Comment, error) {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	r.storage.nextCommentID++

	created := model.Comment{
		ID:        r.storage.nextCommentID,
		UserUUID:  comment.UserUUID,
		PostID:    comment.PostID,
		ParentID:  utils.CloneInt64Ptr(comment.ParentID),
		Content:   comment.Content,
		CreatedAt: time.Now().UTC(),
	}

	r.storage.comments[created.ID] = created

	key := r.storage.makeCommentBucketKey(created.PostID, created.ParentID)
	r.storage.commentsByBucket[key] = append(r.storage.commentsByBucket[key], created.ID)

	return created, nil
}

func (r *CommentRepository) GetByID(_ context.Context, id int64) (model.Comment, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	comment, ok := r.storage.comments[id]
	if !ok {
		return model.Comment{}, coreerrors.ErrCommentNotFound
	}

	comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)

	return comment, nil
}

func (r *CommentRepository) ListByPostAndParent(_ context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
	r.storage.mu.RLock()
	defer r.storage.mu.RUnlock()

	key := r.storage.makeCommentBucketKey(filter.PostID, filter.ParentID)
	bucket := r.storage.commentsByBucket[key]

	comments := make([]model.Comment, 0, len(bucket))
	for _, commentID := range bucket {
		comment, ok := r.storage.comments[commentID]
		if !ok {
			continue
		}

		comment.ParentID = utils.CloneInt64Ptr(comment.ParentID)

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

	return comments[startIdx:endIdx], hasNextPage, nil
}
