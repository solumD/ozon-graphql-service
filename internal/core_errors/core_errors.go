package coreerrors

import "errors"

var (
	ErrPostNotFound               = errors.New("post not found")
	ErrCreatePost                 = errors.New("failed to create post")
	ErrGetPost                    = errors.New("failed to get post")
	ErrListPosts                  = errors.New("failed to list posts")
	ErrUpdateCommentsAvailability = errors.New("failed to update comments availability")
	ErrEmptyPostTitle             = errors.New("post title must not be empty")
	ErrEmptyPostContent           = errors.New("post content must not be empty")
	ErrCommentsDisabled           = errors.New("comments are disabled for this post")

	ErrCommentNotFound     = errors.New("comment not found")
	ErrCreateComment       = errors.New("failed to create comment")
	ErrGetComment          = errors.New("failed to get comment")
	ErrListComments        = errors.New("failed to list comments")
	ErrCommentTooLong      = errors.New("comment is too long")
	ErrEmptyCommentContent = errors.New("comment content must not be empty")

	ErrInvalidParentComment = errors.New("parent comment is invalid")
	ErrInvalidPagination    = errors.New("invalid pagination params")
	ErrInvalidCursor        = errors.New("invalid pagination cursor")
)
