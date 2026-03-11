package models

import "time"

type Sub struct {
	ID          uint64
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
	ID     uint64 `gorm:"primaryKey;pk:user_id,sub_id"`
	User   User   `gorm:"constraint:OnDelete:CASCADE;"`
	UserID uint64
	Sub    Sub `gorm:"constraint:OnDelete:CASCADE;"`
	SubID  uint64
}
