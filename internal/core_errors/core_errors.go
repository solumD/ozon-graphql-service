package coreerrors

import "errors"

var (
	ErrPostNotFound         = errors.New("post not found")
	ErrCommentNotFound      = errors.New("comment not found")
	ErrCommentsDisabled     = errors.New("comments are disabled for this post")
	ErrCommentTooLong       = errors.New("comment is too long")
	ErrInvalidParentComment = errors.New("parent comment is invalid")
	ErrInvalidPagination    = errors.New("invalid pagination params")
	ErrInvalidCursor        = errors.New("invalid pagination cursor")
	ErrEmptyPostTitle       = errors.New("post title must not be empty")
	ErrEmptyPostContent     = errors.New("post content must not be empty")
	ErrEmptyCommentContent  = errors.New("comment content must not be empty")
)
