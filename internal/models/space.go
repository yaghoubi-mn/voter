package models

import "time"

type Space struct {
	ID          uint64
	Username    string `gorm:"size:50;not null;unique"`
	Title       string `gorm:"size:50;not null"`
	Description string `gorm:"size:500;not null"`

	CreatedAt    time.Time `gorm:"autoCreateTime"`
	ModifiedAt   time.Time `gorm:"autoUpdateTime"`
	ClosedByRole string

	Owner   User
	OwnerID uint64

	Views            uint64 `gorm:"default:0"`
	SubscribersCount uint64 `gorm:"default:0"`
}

type Subscription struct {
	User    User   `gorm:"constraint:OnDelete:CASCADE;"`
	UserID  uint64 `gorm:"primaryKey"`
	Space   Space  `gorm:"constraint:OnDelete:CASCADE;"`
	SpaceID uint64 `gorm:"primaryKey"`
}
