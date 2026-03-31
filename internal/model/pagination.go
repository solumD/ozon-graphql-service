package model

import "time"

const (
	DefaultCommentsPageSize = 20
	MaxCommentsPageSize     = 100
	MaxCommentLength        = 2000
)

type Cursor struct {
	CreatedAt time.Time
	ID        int64
}

type CommentListFilter struct {
	PostID   int64
	ParentID *int64
	First    int
	After    *Cursor
}

type CommentEdge struct {
	Cursor string
	Node   Comment
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   *string
}

type CommentConnection struct {
	Edges    []CommentEdge
	PageInfo PageInfo
}
