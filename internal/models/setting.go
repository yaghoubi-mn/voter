package models

import "time"

type Setting struct {
	Key         string `gorm:"primaryKey"`
	Value       string `gorm:"not null"`
	Name        string
	Description string
	ModifiedAt  time.Time

	ModifiedByUser   User
	ModifiedByUserID uint64
}
