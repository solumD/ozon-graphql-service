package tests

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	coreerrors "github.com/solumD/ozon-graphql-service/internal/core_errors"
	"github.com/solumD/ozon-graphql-service/internal/model"
	"github.com/solumD/ozon-graphql-service/internal/usecase"
	"github.com/solumD/ozon-graphql-service/internal/usecase/mock"
	"github.com/solumD/ozon-graphql-service/pkg/logger"
)

func TestCreateComment(t *testing.T) {
	t.Parallel()

	parentID := int64(7)

	tests := []struct {
		name              string
		userUUID          string
		postID            int64
		parentID          *int64
		content           string
		setupPostRepo     func(postRepo *mock.PostRepositoryMock)
		setupCommentRepo  func(commentRepo *mock.CommentRepositoryMock)
		expectedComment   model.Comment
		expectedError     error
		expectedPostCalls uint64
		setupProducer     func(producer *mock.CommentProducerMock)
		expectedPublishes uint64
	}{
		{
			name:          "empty content",
			userUUID:      "user-1",
			postID:        1,
			content:       "   ",
			expectedError: coreerrors.ErrEmptyCommentContent,
		},
		{
			name:          "comment too long",
			userUUID:      "user-1",
			postID:        1,
			content:       strings.Repeat("a", model.MaxCommentLength+1),
			expectedError: coreerrors.ErrCommentTooLong,
		},
		{
			name:     "comment with max allowed length",
			userUUID: "user-1",
			postID:   1,
			content:  strings.Repeat("a", model.MaxCommentLength),
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{ID: 1, CommentsEnabled: true}, nil)
			},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.CreateMock.Set(func(_ context.Context, comment model.Comment) (model.Comment, error) {
					return model.Comment{ID: 100, UserUUID: comment.UserUUID, PostID: comment.PostID, Content: comment.Content}, nil
				})
			},
			expectedComment:   model.Comment{ID: 100, UserUUID: "user-1", PostID: 1, Content: strings.Repeat("a", model.MaxCommentLength)},
			expectedPostCalls: 1,
			setupProducer: func(producer *mock.CommentProducerMock) {
				producer.PublishCommentMock.Expect(minimock.AnyContext, model.Comment{ID: 100, UserUUID: "user-1", PostID: 1, Content: strings.Repeat("a", model.MaxCommentLength)}).Return()
			},
			expectedPublishes: 1,
		},
		{
			name:     "post repository error",
			userUUID: "user-1",
			postID:   2,
			content:  "comment",
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{}, coreerrors.ErrPostNotFound)
			},
			expectedError:     coreerrors.ErrPostNotFound,
			expectedPostCalls: 1,
		},
		{
			name:     "comments disabled",
			userUUID: "user-1",
			postID:   3,
			content:  "comment",
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{ID: 3, CommentsEnabled: false}, nil)
			},
			expectedError:     coreerrors.ErrCommentsDisabled,
			expectedPostCalls: 1,
		},
		{
			name:     "invalid parent comment",
			userUUID: "user-1",
			postID:   4,
			parentID: &parentID,
			content:  "comment",
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{ID: 4, CommentsEnabled: true}, nil)
			},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.GetByIDMock.Return(model.Comment{ID: parentID, PostID: 999}, nil)
			},
			expectedError:     coreerrors.ErrInvalidParentComment,
			expectedPostCalls: 1,
		},
		{
			name:     "parent lookup error",
			userUUID: "user-1",
			postID:   4,
			parentID: &parentID,
			content:  "comment",
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{ID: 4, CommentsEnabled: true}, nil)
			},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.GetByIDMock.Return(model.Comment{}, coreerrors.ErrCommentNotFound)
			},
			expectedError:     coreerrors.ErrCommentNotFound,
			expectedPostCalls: 1,
		},
		{
			name:     "success without parent",
			userUUID: " user-1 ",
			postID:   5,
			content:  " content ",
			setupPostRepo: func(postRepo *mock.PostRepositoryMock) {
				postRepo.GetByIDMock.Return(model.Post{ID: 5, CommentsEnabled: true}, nil)
			},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.CreateMock.Set(func(_ context.Context, comment model.Comment) (model.Comment, error) {
					if comment.UserUUID != "user-1" {
						t.Fatalf("unexpected userUUID: %q", comment.UserUUID)
					}
					if comment.Content != "content" {
						t.Fatalf("unexpected content: %q", comment.Content)
					}
					if comment.ParentID != nil {
						t.Fatal("expected nil parentID")
					}

					return model.Comment{ID: 101, UserUUID: comment.UserUUID, PostID: comment.PostID, Content: comment.Content}, nil
				})
			},
			expectedComment:   model.Comment{ID: 101, UserUUID: "user-1", PostID: 5, Content: "content"},
			expectedPostCalls: 1,
			setupProducer: func(producer *mock.CommentProducerMock) {
				producer.PublishCommentMock.Expect(minimock.AnyContext, model.Comment{ID: 101, UserUUID: "user-1", PostID: 5, Content: "content"}).Return()
			},
			expectedPublishes: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			commentRepo := mock.NewCommentRepositoryMock(mc)
			producer := mock.NewCommentProducerMock(mc)

			if tt.setupPostRepo != nil {
				tt.setupPostRepo(postRepo)
			}
			if tt.setupCommentRepo != nil {
				tt.setupCommentRepo(commentRepo)
			}
			if tt.setupProducer != nil {
				tt.setupProducer(producer)
			}

			uc := usecase.NewCommentUsecase(postRepo, commentRepo, producer, logger.NewLogger("error"))
			comment, err := uc.CreateComment(context.Background(), tt.userUUID, tt.postID, tt.parentID, tt.content)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError == nil && comment != tt.expectedComment {
				t.Fatalf("expected comment %+v, got %+v", tt.expectedComment, comment)
			}

			if got := postRepo.GetByIDBeforeCounter(); got != tt.expectedPostCalls {
				t.Fatalf("expected GetByID calls %d, got %d", tt.expectedPostCalls, got)
			}

			if tt.parentID == nil {
				if got := commentRepo.GetByIDBeforeCounter(); got != 0 {
					t.Fatalf("expected parent comment lookup not to be called, got %d", got)
				}
			}

			if got := producer.PublishCommentBeforeCounter(); got != tt.expectedPublishes {
				t.Fatalf("expected PublishComment calls %d, got %d", tt.expectedPublishes, got)
			}
		})
	}
}

func TestListComments(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	comments := []model.Comment{
		{ID: 1, PostID: 10, Content: "first", CreatedAt: now},
		{ID: 2, PostID: 10, Content: "second", CreatedAt: now.Add(time.Second)},
	}

	tests := []struct {
		name              string
		filter            model.CommentListFilter
		setupCommentRepo  func(commentRepo *mock.CommentRepositoryMock)
		expectedError     error
		expectedEdges     int
		expectedHasNext   bool
		expectedFirstNorm int
	}{
		{
			name:          "invalid post id",
			filter:        model.CommentListFilter{PostID: 0},
			expectedError: coreerrors.ErrInvalidPagination,
		},
		{
			name:   "success with default page size",
			filter: model.CommentListFilter{PostID: 10},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.ListByPostAndParentMock.Set(func(_ context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
					if filter.First != model.DefaultCommentsPageSize {
						t.Fatalf("expected normalized First=%d, got %d", model.DefaultCommentsPageSize, filter.First)
					}
					return comments, true, nil
				})
			},
			expectedEdges:   2,
			expectedHasNext: true,
		},
		{
			name:   "success with clamped page size",
			filter: model.CommentListFilter{PostID: 10, First: model.MaxCommentsPageSize + 100},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.ListByPostAndParentMock.Set(func(_ context.Context, filter model.CommentListFilter) ([]model.Comment, bool, error) {
					if filter.First != model.MaxCommentsPageSize {
						t.Fatalf("expected normalized First=%d, got %d", model.MaxCommentsPageSize, filter.First)
					}
					return comments[:1], false, nil
				})
			},
			expectedEdges:   1,
			expectedHasNext: false,
		},
		{
			name:   "repository error",
			filter: model.CommentListFilter{PostID: 10, First: 5},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.ListByPostAndParentMock.Return(nil, false, coreerrors.ErrListComments)
			},
			expectedError: coreerrors.ErrListComments,
		},
		{
			name:   "empty list",
			filter: model.CommentListFilter{PostID: 10},
			setupCommentRepo: func(commentRepo *mock.CommentRepositoryMock) {
				commentRepo.ListByPostAndParentMock.Return([]model.Comment{}, false, nil)
			},
			expectedEdges:   0,
			expectedHasNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mc := minimock.NewController(t)
			postRepo := mock.NewPostRepositoryMock(mc)
			commentRepo := mock.NewCommentRepositoryMock(mc)
			if tt.setupCommentRepo != nil {
				tt.setupCommentRepo(commentRepo)
			}

			uc := usecase.NewCommentUsecase(postRepo, commentRepo, nil, logger.NewLogger("error"))
			connection, err := uc.ListComments(context.Background(), tt.filter)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}

			if tt.expectedError != nil {
				return
			}

			if len(connection.Edges) != tt.expectedEdges {
				t.Fatalf("expected %d edges, got %d", tt.expectedEdges, len(connection.Edges))
			}

			if connection.PageInfo.HasNextPage != tt.expectedHasNext {
				t.Fatalf("expected hasNextPage=%v, got %v", tt.expectedHasNext, connection.PageInfo.HasNextPage)
			}

			if len(connection.Edges) > 0 {
				if connection.PageInfo.EndCursor == nil {
					t.Fatal("expected end cursor to be set")
				}

				cursor, err := usecase.DecodeCursor(*connection.PageInfo.EndCursor)
				if err != nil {
					t.Fatalf("failed to decode end cursor: %v", err)
				}

				last := connection.Edges[len(connection.Edges)-1].Node
				if cursor.ID != last.ID {
					t.Fatalf("expected cursor ID %d, got %d", last.ID, cursor.ID)
				}
			} else if connection.PageInfo.EndCursor != nil {
				t.Fatalf("expected nil end cursor for empty edges, got %v", *connection.PageInfo.EndCursor)
			}
		})
	}
}
