package models

import "time"

type Comment struct {
	ID      uint64
	Content string `gorm:"size:300;not null"`

	Score      int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ModifiedAt time.Time `gorm:"autoUpdateTime:false"`
	IsDeleted  bool      `gorm:"default:false"`

	ParentID uint64
	AuthorID uint64
	Author   User `gorm:"constraint:OnDelete:CASCADE;"`
	PostID   uint64
	Post     Post `gorm:"constraint:OnDelete:CASCADE;"`
}

type CommentVote struct {
	ID        uint64  `gorm:"primaryKey;pk:comment_id,user_id"`
	Comment   Comment `gorm:"constraint:OnDelete:CASCADE;"`
	CommentID uint64

	UserID uint64
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

	Vote bool // true for up vote and false for down vote
}
