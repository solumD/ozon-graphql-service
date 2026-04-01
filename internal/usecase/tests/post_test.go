package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	coreerrors "github.com/solumD/ozon-grapql-service/internal/core_errors"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/usecase"
	"github.com/solumD/ozon-grapql-service/internal/usecase/mock"
	"github.com/solumD/ozon-grapql-service/pkg/logger"
)

func TestСreatePost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userUUID      string
		title         string
		content       string
		commentsOn    bool
		setupMock     func(postRepo *mock.PostRepositoryMock)
		expectedPost  model.Post
		expectedError error
		assertCalls   func(t *testing.T, postRepo *mock.PostRepositoryMock)
	}{
		{
			name:       "success with trim",
			userUUID:   " user-1 ",
			title:      " title ",
			content:    " content ",
			commentsOn: true,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.CreateMock.Set(func(_ context.Context, post model.Post) (model.Post, error) {
					if post.UserUUID != "user-1" {
						t.Fatalf("unexpected userUUID: %q", post.UserUUID)
					}
					if post.Title != "title" {
						t.Fatalf("unexpected title: %q", post.Title)
					}
					if post.Content != "content" {
						t.Fatalf("unexpected content: %q", post.Content)
					}
					if !post.CommentsEnabled {
						t.Fatal("expected comments enabled")
					}

					return model.Post{ID: 1, UserUUID: post.UserUUID, Title: post.Title, Content: post.Content, CommentsEnabled: post.CommentsEnabled}, nil
				})
			},
			expectedPost: model.Post{ID: 1, UserUUID: "user-1", Title: "title", Content: "content", CommentsEnabled: true},
			assertCalls: func(t *testing.T, postRepo *mock.PostRepositoryMock) {
				t.Helper()
				if got := postRepo.CreateBeforeCounter(); got != 1 {
					t.Fatalf("expected Create to be called once, got %d", got)
				}
			},
		},
		{
			name:          "empty title",
			userUUID:      "user-1",
			title:         "   ",
			content:       "content",
			expectedError: coreerrors.ErrEmptyPostTitle,
			assertCalls: func(t *testing.T, postRepo *mock.PostRepositoryMock) {
				t.Helper()
				if got := postRepo.CreateBeforeCounter(); got != 0 {
					t.Fatalf("expected Create not to be called, got %d", got)
				}
			},
		},
		{
			name:          "empty content",
			userUUID:      "user-1",
			title:         "title",
			content:       "   ",
			expectedError: coreerrors.ErrEmptyPostContent,
			assertCalls: func(t *testing.T, postRepo *mock.PostRepositoryMock) {
				t.Helper()
				if got := postRepo.CreateBeforeCounter(); got != 0 {
					t.Fatalf("expected Create not to be called, got %d", got)
				}
			},
		},
		{
			name:       "repository error",
			userUUID:   "user-1",
			title:      "title",
			content:    "content",
			commentsOn: true,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.CreateMock.Return(model.Post{}, coreerrors.ErrCreatePost)
			},
			expectedError: coreerrors.ErrCreatePost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			if tt.setupMock != nil {
				tt.setupMock(postRepo)
			}

			uc := usecase.NewPostUsecase(postRepo, logger.NewLogger("error"))
			post, err := uc.CreatePost(context.Background(), tt.userUUID, tt.title, tt.content, tt.commentsOn)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil && post != tt.expectedPost {
				t.Fatalf("expected post %+v, got %+v", tt.expectedPost, post)
			}

			if tt.assertCalls != nil {
				tt.assertCalls(t, postRepo)
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		postID        int64
		setupMock     func(postRepo *mock.PostRepositoryMock)
		expectedPost  model.Post
		expectedError error
	}{
		{
			name:   "success",
			postID: 10,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Set(func(_ context.Context, id int64) (model.Post, error) {
					if id != 10 {
						t.Fatalf("unexpected postID: %d", id)
					}

					return model.Post{ID: 10, Title: "title"}, nil
				})
			},
			expectedPost: model.Post{ID: 10, Title: "title"},
		},
		{
			name:   "repository error",
			postID: 11,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{}, coreerrors.ErrPostNotFound)
			},
			expectedError: coreerrors.ErrPostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			tt.setupMock(postRepo)

			uc := usecase.NewPostUsecase(postRepo, logger.NewLogger("error"))
			post, err := uc.GetPost(context.Background(), tt.postID)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil && post != tt.expectedPost {
				t.Fatalf("expected post %+v, got %+v", tt.expectedPost, post)
			}
		})
	}
}

func TestListPosts(t *testing.T) {
	t.Parallel()

	expectedPosts := []model.Post{{ID: 1, Title: "first"}, {ID: 2, Title: "second"}}

	tests := []struct {
		name          string
		setupMock     func(postRepo *mock.PostRepositoryMock)
		expectedPosts []model.Post
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.ListMock.Return(expectedPosts, nil)
			},
			expectedPosts: expectedPosts,
		},
		{
			name: "repository error",
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.ListMock.Return(nil, coreerrors.ErrListPosts)
			},
			expectedError: coreerrors.ErrListPosts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			tt.setupMock(postRepo)

			uc := usecase.NewPostUsecase(postRepo, logger.NewLogger("error"))
			posts, err := uc.ListPosts(context.Background())

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil && len(posts) != len(tt.expectedPosts) {
				t.Fatalf("expected %d posts, got %d", len(tt.expectedPosts), len(posts))
			}
		})
	}
}

func TestChangeCommentsAvailability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		postID        int64
		enabled       bool
		setupMock     func(postRepo *mock.PostRepositoryMock)
		expectedPost  model.Post
		expectedError error
	}{
		{
			name:    "success",
			postID:  15,
			enabled: false,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.UpdateCommentsAvailabilityMock.Set(func(_ context.Context, postID int64, enabled bool) (model.Post, error) {
					if postID != 15 || enabled != false {
						t.Fatalf("unexpected args: postID=%d enabled=%v", postID, enabled)
					}

					return model.Post{ID: 15, CommentsEnabled: false}, nil
				})
			},
			expectedPost: model.Post{ID: 15, CommentsEnabled: false},
		},
		{
			name:    "repository error",
			postID:  15,
			enabled: true,
			setupMock: func(postRepo *mock.PostRepositoryMock) {
				postRepo.UpdateCommentsAvailabilityMock.Return(model.Post{}, coreerrors.ErrUpdateCommentsAvailability)
			},
			expectedError: coreerrors.ErrUpdateCommentsAvailability,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			tt.setupMock(postRepo)

			uc := usecase.NewPostUsecase(postRepo, logger.NewLogger("error"))
			post, err := uc.ChangeCommentsAvailability(context.Background(), tt.postID, tt.enabled)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil && post != tt.expectedPost {
				t.Fatalf("expected post %+v, got %+v", tt.expectedPost, post)
			}
		})
	}
}
