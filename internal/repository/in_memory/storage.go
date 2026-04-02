package inmemory

import (
	"fmt"
	"sync"

	"github.com/solumD/ozon-grapql-service/internal/model"
)

type Storage struct {
	mu sync.RWMutex

	nextPostID    int64
	nextCommentID int64

	posts            map[int64]model.Post
	comments         map[int64]model.Comment
	commentsByBucket map[string][]int64
}

func NewStorage() *Storage {
	return &Storage{
		posts:            make(map[int64]model.Post),
		comments:         make(map[int64]model.Comment),
		commentsByBucket: make(map[string][]int64),
	}
}

func (s *Storage) makeCommentBucketKey(postID int64, parentID *int64) string {
	if parentID == nil {
		return fmt.Sprintf("%d:root", postID)
	}

	return fmt.Sprintf("%d:%d", postID, *parentID)
}
