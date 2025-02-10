package models

import "time"

type Post struct {
	ID      uint64
	Title   string `gorm:"size:200;not null"`
	Content string `gorm:"size:1000;not null"`

	Score      int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ModifiedAt time.Time `gorm:"autoUpdateTime"`

	AuthorID uint64
	Author   User
}

type PostVote struct {
	ID uint64

	PostID uint64
	Post   Post

	UserID uint64
	User   User
}
