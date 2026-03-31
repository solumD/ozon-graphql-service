package graphql

import (
	"github.com/solumD/ozon-grapql-service/internal/delivery/graphql/generated"
	"github.com/solumD/ozon-grapql-service/internal/model"
	"github.com/solumD/ozon-grapql-service/internal/utils"
)

func toGraphQLPost(post model.Post) *generated.Post {
	return &generated.Post{
		ID:              int(post.ID),
		UserUUID:        post.UserUUID,
		Title:           post.Title,
		Content:         post.Content,
		CommentsEnabled: post.CommentsEnabled,
		CreatedAt:       post.CreatedAt,
	}
}

func toGraphQLComment(comment model.Comment) *generated.Comment {
	return &generated.Comment{
		ID:         int(comment.ID),
		UserUUID:   comment.UserUUID,
		PostID:     int(comment.PostID),
		ParentID:   utils.ToIntPtr(comment.ParentID),
		HasReplies: comment.HasReplies,
		Content:    comment.Content,
		CreatedAt:  comment.CreatedAt,
	}
}

func toGraphQLCommentConnection(connection model.CommentConnection) *generated.CommentConnection {
	edges := make([]*generated.CommentEdge, 0, len(connection.Edges))
	for _, edge := range connection.Edges {
		edges = append(edges, &generated.CommentEdge{
			Cursor: edge.Cursor,
			Node:   toGraphQLComment(edge.Node),
		})
	}

	return &generated.CommentConnection{
		Edges: edges,
		PageInfo: &generated.PageInfo{
			HasNextPage: connection.PageInfo.HasNextPage,
			EndCursor:   connection.PageInfo.EndCursor,
		},
	}
}
