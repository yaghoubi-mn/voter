package models

import "time"

type Comment struct {
	ID      uint64
	Content string `gorm:"size:1000;not null"`

	Score      int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ModifiedAt time.Time `gorm:"autoUpdateTime"`

	CommentID uint64
	AuthorID  uint64
	Author    User `gorm:"constraint:OnDelete:CASCADE;"`
	PostID    uint64
	Post      Post `gorm:"constraint:OnDelete:CASCADE;"`
}

type CommentVote struct {
	ID        uint64
	CommentID uint64
	Comment   Comment `gorm:"constraint:OnDelete:CASCADE;"`

	UserID uint64
	User   User `gorm:"constraint:OnDelete:CASCADE;"`
}
