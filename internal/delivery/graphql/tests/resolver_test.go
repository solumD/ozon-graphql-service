package tests

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	graphql "github.com/solumD/ozon-grapql-service/internal/delivery/graphql"
	"github.com/solumD/ozon-grapql-service/internal/delivery/graphql/mock"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
)

func encodeTestCursor(cursor model.Cursor) string {
	raw := fmt.Sprintf("%d:%d", cursor.CreatedAt.UnixNano(), cursor.ID)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func TestCreatePost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userUUID      string
		title         string
		content       string
		commentsOn    bool
		setupMock     func(postUC *mock.PostUsecaseMock)
		expectedError error
		expectedID    int
		expectedUser  string
	}{
		{
			name:       "success",
			userUUID:   "user-1",
			title:      "title",
			content:    "content",
			commentsOn: true,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.CreatePostMock.Expect(context.Background(), "user-1", "title", "content", true).
					Return(model.Post{ID: 1, UserUUID: "user-1", Title: "title", Content: "content", CommentsEnabled: true}, nil)
			},
			expectedID:   1,
			expectedUser: "user-1",
		},
		{
			name:       "usecase error",
			userUUID:   "user-1",
			title:      "title",
			content:    "content",
			commentsOn: true,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.CreatePostMock.Expect(context.Background(), "user-1", "title", "content", true).
					Return(model.Post{}, coreerrors.ErrCreatePost)
			},
			expectedError: coreerrors.ErrCreatePost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			if tt.setupMock != nil {
				tt.setupMock(postUC)
			}

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Mutation().CreatePost(context.Background(), tt.userUUID, tt.title, tt.content, tt.commentsOn)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil {
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				if result.ID != tt.expectedID {
					t.Fatalf("expected id=%d, got %d", tt.expectedID, result.ID)
				}
				if result.UserUUID != tt.expectedUser {
					t.Fatalf("expected userUUID=%s, got %s", tt.expectedUser, result.UserUUID)
				}
			}
		})
	}
}

func TestChangePostCommentsAvailability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		postID        int
		enabled       bool
		setupMock     func(postUC *mock.PostUsecaseMock)
		expectedError error
		expectedID    int
		expectedFlag  bool
	}{
		{
			name:    "success",
			postID:  5,
			enabled: true,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.ChangeCommentsAvailabilityMock.Expect(context.Background(), int64(5), true).
					Return(model.Post{ID: 5, CommentsEnabled: true}, nil)
			},
			expectedID:   5,
			expectedFlag: true,
		},
		{
			name:    "usecase error",
			postID:  5,
			enabled: true,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.ChangeCommentsAvailabilityMock.Expect(context.Background(), int64(5), true).
					Return(model.Post{}, coreerrors.ErrPostNotFound)
			},
			expectedError: coreerrors.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			tt.setupMock(postUC)

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Mutation().ChangePostCommentsAvailability(context.Background(), tt.postID, tt.enabled)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil {
				if result.ID != tt.expectedID || result.CommentsEnabled != tt.expectedFlag {
					t.Fatalf("unexpected result: %+v", result)
				}
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	t.Parallel()

	parentID := 7
	parentID64 := int64(7)

	tests := []struct {
		name           string
		userUUID       string
		postID         int
		parentID       *int
		content        string
		setupMock      func(commentUC *mock.CommentUsecaseMock)
		expectedError  error
		expectedID     int
		expectedParent *int
	}{
		{
			name:     "success",
			userUUID: "user-1",
			postID:   3,
			parentID: &parentID,
			content:  "hello",
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				commentUC.CreateCommentMock.Expect(context.Background(), "user-1", int64(3), &parentID64, "hello").
					Return(model.Comment{ID: 11, UserUUID: "user-1", PostID: 3, ParentID: &parentID64, Content: "hello"}, nil)
			},
			expectedID:     11,
			expectedParent: &parentID,
		},
		{
			name:     "usecase error",
			userUUID: "user-1",
			postID:   3,
			content:  "hello",
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				commentUC.CreateCommentMock.Expect(context.Background(), "user-1", int64(3), (*int64)(nil), "hello").
					Return(model.Comment{}, coreerrors.ErrCreateComment)
			},
			expectedError: coreerrors.ErrCreateComment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			tt.setupMock(commentUC)

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Mutation().CreateComment(context.Background(), tt.userUUID, tt.postID, tt.parentID, tt.content)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil {
				if result == nil || result.ID != tt.expectedID {
					t.Fatalf("unexpected result: %+v", result)
				}
				if tt.expectedParent == nil {
					if result.ParentID != nil {
						t.Fatalf("expected nil parentID, got %#v", result.ParentID)
					}
				} else if result.ParentID == nil || *result.ParentID != *tt.expectedParent {
					t.Fatalf("unexpected parentID: %#v", result.ParentID)
				}
			}
		})
	}
}

func TestPosts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMock     func(postUC *mock.PostUsecaseMock)
		expectedError error
		expectedCount int
	}{
		{
			name: "success",
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.ListPostsMock.Return([]model.Post{{ID: 1}, {ID: 2}}, nil)
			},
			expectedCount: 2,
		},
		{
			name: "usecase error",
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.ListPostsMock.Return(nil, coreerrors.ErrListPosts)
			},
			expectedError: coreerrors.ErrListPosts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			tt.setupMock(postUC)

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Query().Posts(context.Background())

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil && len(result) != tt.expectedCount {
				t.Fatalf("expected count %d, got %d", tt.expectedCount, len(result))
			}
		})
	}
}

func TestPost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            int
		setupMock     func(postUC *mock.PostUsecaseMock)
		expectedError error
		expectedID    int
	}{
		{
			name: "success",
			id:   9,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.GetPostMock.Expect(context.Background(), int64(9)).Return(model.Post{ID: 9, Title: "post"}, nil)
			},
			expectedID: 9,
		},
		{
			name: "usecase error",
			id:   9,
			setupMock: func(postUC *mock.PostUsecaseMock) {
				postUC.GetPostMock.Expect(context.Background(), int64(9)).Return(model.Post{}, coreerrors.ErrPostNotFound)
			},
			expectedError: coreerrors.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			tt.setupMock(postUC)

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Query().Post(context.Background(), tt.id)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil && result.ID != tt.expectedID {
				t.Fatalf("expected id %d, got %d", tt.expectedID, result.ID)
			}
		})
	}
}

func TestComments(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	cursor := encodeTestCursor(model.Cursor{CreatedAt: now, ID: 10})
	invalidCursor := "%%%"

	tests := []struct {
		name           string
		postID         int
		parentID       *int
		first          *int
		after          *string
		setupMock      func(commentUC *mock.CommentUsecaseMock)
		expectedError  error
		expectedCount  int
		expectedCursor bool
	}{
		{
			name:          "invalid cursor",
			postID:        5,
			after:         &invalidCursor,
			expectedError: coreerrors.ErrInvalidCursor,
		},
		{
			name:   "success",
			postID: 5,
			first:  func() *int { v := 2; return &v }(),
			after:  &cursor,
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				expectedFilter := model.CommentListFilter{PostID: 5, First: 2, After: &model.Cursor{CreatedAt: now, ID: 10}}
				commentUC.ListCommentsMock.Expect(context.Background(), expectedFilter).Return(
					model.CommentConnection{
						Edges:    []model.CommentEdge{{Cursor: "c1", Node: model.Comment{ID: 1}}, {Cursor: "c2", Node: model.Comment{ID: 2}}},
						PageInfo: model.PageInfo{HasNextPage: true, EndCursor: func() *string { s := "c2"; return &s }()},
					},
					nil,
				)
			},
			expectedCount:  2,
			expectedCursor: true,
		},
		{
			name:   "success with parent id",
			postID: 5,
			parentID: func() *int {
				v := 42
				return &v
			}(),
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				parentID64 := int64(42)
				expectedFilter := model.CommentListFilter{PostID: 5, ParentID: &parentID64}
				commentUC.ListCommentsMock.Expect(context.Background(), expectedFilter).Return(
					model.CommentConnection{Edges: []model.CommentEdge{}, PageInfo: model.PageInfo{HasNextPage: false, EndCursor: nil}},
					nil,
				)
			},
			expectedCount:  0,
			expectedCursor: false,
		},
		{
			name:   "success empty response",
			postID: 5,
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				expectedFilter := model.CommentListFilter{PostID: 5}
				commentUC.ListCommentsMock.Expect(context.Background(), expectedFilter).Return(
					model.CommentConnection{Edges: []model.CommentEdge{}, PageInfo: model.PageInfo{HasNextPage: false, EndCursor: nil}},
					nil,
				)
			},
			expectedCount:  0,
			expectedCursor: false,
		},
		{
			name:   "usecase error",
			postID: 5,
			setupMock: func(commentUC *mock.CommentUsecaseMock) {
				expectedFilter := model.CommentListFilter{PostID: 5}
				commentUC.ListCommentsMock.Expect(context.Background(), expectedFilter).Return(model.CommentConnection{}, coreerrors.ErrListComments)
			},
			expectedError: coreerrors.ErrListComments,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mc := minimock.NewController(t)
			postUC := mock.NewPostUsecaseMock(mc)
			commentUC := mock.NewCommentUsecaseMock(mc)
			if tt.setupMock != nil {
				tt.setupMock(commentUC)
			}

			consumer := mock.NewCommentConsumerMock(mc)
			resolver := graphql.NewResolver(postUC, commentUC, consumer, logger.NewLogger("error"))
			result, err := resolver.Query().Comments(context.Background(), tt.postID, tt.parentID, tt.first, tt.after)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil {
				if len(result.Edges) != tt.expectedCount {
					t.Fatalf("expected %d edges, got %d", tt.expectedCount, len(result.Edges))
				}
				if tt.expectedCursor && (result.PageInfo == nil || result.PageInfo.EndCursor == nil) {
					t.Fatal("expected end cursor to be set")
				}
			}
		})
	}
}
