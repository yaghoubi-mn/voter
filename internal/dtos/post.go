package dtos

import (
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/models"
)

type PostInput struct {
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type PostOutput struct {
	ID             uint64 `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AuthorID       uint64 `json:"author_id"`
	AuthorUsername string `json:"author_username"`
	SpaceID        uint64 `json:"space_id"`
	SpaceUsername  string `json:"space_username"`
	Score          int    `json:"score"`
	CreatedAt      string `json:"created_at"`
	ModifiedAt     string `json:"modified_at"`
	CommentsCount  int    `json:"comments_count"`
	Views          uint64 `json:"views"`
}

func GetPostOutputFromPost(post models.Post) PostOutput {
	return PostOutput{
		ID:             post.ID,
		Title:          post.Title,
		Content:        post.Content,
		AuthorID:       post.AuthorID,
		AuthorUsername: post.Author.Username,
		SpaceID:        post.Space.ID,
		SpaceUsername:  post.Space.Username,
		Score:          post.Score,
		CreatedAt:      post.CreatedAt.Format(config.TimeFormat),
		ModifiedAt:     post.ModifiedAt.Format(config.TimeFormat),
		CommentsCount:  post.CommentsCount,
		Views:          post.Views,
	}
}
