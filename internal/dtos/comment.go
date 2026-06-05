package dtos

import (
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/models"
)

type CommentInput struct {
	Content  string `json:"content" validate:"required"`
	ParentID uint64 `json:"parent_id"`
}

type CommentOutput struct {
	ID             uint64 `json:"id"`
	Content        string `json:"content"`
	Score          int    `json:"score"`
	CreatedAt      string `json:"created_at"`
	ModifiedAt     string `json:"modified_at"`
	AuthorID       uint64 `json:"author_id"`
	AuthorUsername string `json:"author_username"`
	ParentID       uint64 `json:"parent_id"`
}

func GetCommentOutputFromComment(comment models.Comment) CommentOutput {
	return CommentOutput{
		ID:             comment.ID,
		Content:        comment.Content,
		Score:          comment.Score,
		CreatedAt:      comment.CreatedAt.Format(config.TimeFormat),
		ModifiedAt:     comment.ModifiedAt.Format(config.TimeFormat),
		AuthorID:       comment.AuthorID,
		AuthorUsername: comment.Author.Username,
		ParentID:       comment.ParentID,
	}
}
