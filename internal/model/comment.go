package model

import "time"

type Comment struct {
	ID         int64
	UserUUID   string
	PostID     int64
	ParentID   *int64
	HasReplies bool
	Content    string
	CreatedAt  time.Time
}
