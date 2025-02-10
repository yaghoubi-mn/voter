package models

type User struct {
	ID       uint64
	Username string `gorm:"size:50;not null;unique"`
	Password string `gorm:"size:300;not null"`
	Salt     string `gorm:"size:300;not null"`
}
