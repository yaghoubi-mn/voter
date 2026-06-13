package custom_errors

import (
	"gorm.io/gorm"
)

var (
	RecordNotFound error = gorm.ErrRecordNotFound
	DuplicateKey   error = gorm.ErrDuplicatedKey
)
