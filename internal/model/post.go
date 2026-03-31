package model

import "time"

type Post struct {
	ID              int64
	UserUUID        string
	Title           string
	Content         string
	CommentsEnabled bool
	CreatedAt       time.Time
}
