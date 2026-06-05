package repositories

import (
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetByUsername(username string) (models.User, error)
	Create(user *models.User) error
	Update(user models.User) error
	Delete(userID uint64) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetByUsername(username string) (models.User, error) {

	var user models.User
	if err := r.db.Where(&models.User{Username: username}).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return user, custom_errors.RecordNotFound
		}
		return user, err
	}

	return user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user models.User) error {
	return r.db.Updates(&user).Error

}

func (r *userRepository) Delete(userId uint64) error {
	return r.db.Delete(models.User{ID: userId}).Error
}
