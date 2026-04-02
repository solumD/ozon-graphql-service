package inmemory

import (
	"context"
	"sync"

	"github.com/solumD/ozon-grapql-service/internal/model"
)

type CommentBroker struct {
	mu          sync.RWMutex
	subscribers map[int64]map[chan model.Comment]struct{}
}

func NewCommentBroker() *CommentBroker {
	return &CommentBroker{
		subscribers: make(map[int64]map[chan model.Comment]struct{}),
	}
}

// PublishComment отправляет комментарий всем подписчикам
func (b *CommentBroker) PublishComment(_ context.Context, comment model.Comment) {
	b.mu.RLock()
	subscribers := b.subscribers[comment.PostID]
	channels := make([]chan model.Comment, 0, len(subscribers))
	for ch := range subscribers {
		channels = append(channels, ch)
	}
	b.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- comment:
		default:
		}
	}
}

// SubscribeToComments оформляет подписку на комментарии и возвращает канал для их
func (b *CommentBroker) SubscribeToComments(ctx context.Context, postID int64) <-chan model.Comment {
	ch := make(chan model.Comment, 1)

	b.mu.Lock()
	if _, ok := b.subscribers[postID]; !ok {
		b.subscribers[postID] = make(map[chan model.Comment]struct{})
	}
	b.subscribers[postID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.unsubscribe(postID, ch)
	}()

	return ch
}

// unsubscribe удаляет подписку
func (b *CommentBroker) unsubscribe(postID int64, ch chan model.Comment) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscribers, ok := b.subscribers[postID]
	if !ok {
		close(ch)
		return
	}

	if _, ok := subscribers[ch]; ok {
		delete(subscribers, ch)
		close(ch)
	}

	if len(subscribers) == 0 {
		delete(b.subscribers, postID)
	}
}
