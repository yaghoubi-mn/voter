package repositories

import (
	"fmt"

	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

type SubRepository interface {
	Create(sub models.Sub) error
	Update(sub models.Sub) error
	Delete(subId uint64) error
	GetByID(subId uint64) (models.Sub, error)
	GetAll(sortBy enums.SortBy, page int) ([]models.Sub, error)

	AddSubScore(subId uint64, number int) error
}

type subRepository struct {
	db *gorm.DB
}

func NewSubRepository(db *gorm.DB) SubRepository {
	return &subRepository{
		db: db,
	}
}

func (r *subRepository) Create(sub models.Sub) error {
	return r.db.Create(&sub).Error
}

func (r *subRepository) Update(sub models.Sub) error {
	return r.db.Updates(&sub).Error

}

func (r *subRepository) Delete(subId uint64) error {
	return r.db.Delete(models.Sub{ID: subId}).Error
}

func (r *subRepository) GetByID(subId uint64) (models.Sub, error) {
	var sub models.Sub
	if err := r.db.Preload("Owner").First(&sub, &models.Sub{ID: subId}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return sub, custom_errors.RecordNotFound
		}

		return sub, err
	}

	return sub, nil
}

func (r *subRepository) GetAll(sortBy enums.SortBy, page int) ([]models.Sub, error) {
	var subs []models.Sub
	err := r.db.Preload("Owner").Order(string(sortBy)).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&subs).Error

	return subs, err
}

func (r *subRepository) AddSubScore(subId uint64, number int) error {
	var expr string
	if number >= 0 {
		expr = fmt.Sprintf("score + %v", number)
	} else {
		expr = fmt.Sprintf("score %v", number)
	}

	if err := r.db.Model(&models.Sub{}).Where("id=?", subId).Update("score", gorm.Expr(expr)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return custom_errors.RecordNotFound
		}

		return err
	}

	return nil
}

func (r *subRepository) SubscribeSub(userID uint64, subID uint64) error {
	s := models.Subscription{
		UserID: userID,
		SubID:  subID,
	}
	if err := r.db.Create(&s).Error; err != nil {
		return err
	}

	return nil
}

func (r *subRepository) UnsubscribeSub(userID, subID uint64) error {
	s := models.Subscription{
		UserID: userID,
		SubID:  subID,
	}
	if err := r.db.Model(&models.Subscription{}).Delete(&s, &s).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	return nil
}
