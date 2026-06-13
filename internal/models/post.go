package models

import "time"

type Post struct {
	ID      uint64
	Title   string `gorm:"size:200;not null"`
	Content string `gorm:"size:1000;not null"`

	Score         int       `gorm:"default:0"`
	CommentsCount int       `gorm:"default:0"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	ModifiedAt    time.Time `gorm:"autoUpdateTime:false"`

	AuthorID uint64
	Author   User `gorm:"constraint:OnDelete:CASCADE;"`

	SpaceID uint64
	Space   Space `gorm:"constraint:OnDelete:CASCADE;"`
}

type PostVote struct {
	ID uint64

	PostID uint64
	Post   Post `gorm:"constraint:OnDelete:CASCADE;"`

	UserID uint64
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

	Vote bool // true for up vote and false for down vote
}
