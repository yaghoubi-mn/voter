package repositories

import (
	"fmt"

	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

type SpaceRepository interface {
	Create(space *models.Space) error
	Update(space models.Space) error
	Delete(spaceId uint64) error
	GetByID(spaceId uint64) (models.Space, error)
	GetAll(sortBy enums.SortBy, page int) ([]models.Space, error)

	AddSubScore(spaceId uint64, number int) error
	SubscribeSub(userID, spaceID uint64) error
	UnsubscribeSub(userID, spaceID uint64) error
	GetUserSubscriptions(userID uint64) ([]models.Space, error)
	IncreaseSpaceSubscribersCount(spaceId uint64, count int) error
}

type spaceRepository struct {
	db *gorm.DB
}

func NewSubRepository(db *gorm.DB) SpaceRepository {
	return &spaceRepository{
		db: db,
	}
}

func (r *spaceRepository) Create(space *models.Space) error {
	return r.db.Create(space).Error
}

func (r *spaceRepository) Update(space models.Space) error {
	return r.db.Updates(&space).Error

}

func (r *spaceRepository) Delete(spaceId uint64) error {
	return r.db.Delete(models.Space{ID: spaceId}).Error
}

func (r *spaceRepository) GetByID(spaceId uint64) (models.Space, error) {
	var space models.Space
	if err := r.db.Preload("Owner").First(&space, &models.Space{ID: spaceId}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return space, custom_errors.RecordNotFound
		}

		return space, err
	}

	return space, nil
}

func (r *spaceRepository) GetAll(sortBy enums.SortBy, page int) ([]models.Space, error) {
	var spaces []models.Space
	err := r.db.Preload("Owner").Order(string(sortBy)).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&spaces).Error

	return spaces, err
}

func (r *spaceRepository) AddSubScore(spaceId uint64, number int) error {
	var expr string
	if number >= 0 {
		expr = fmt.Sprintf("score + %v", number)
	} else {
		expr = fmt.Sprintf("score %v", number)
	}

	if err := r.db.Model(&models.Space{}).Where("id=?", spaceId).Update("score", gorm.Expr(expr)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return custom_errors.RecordNotFound
		}

		return err
	}

	return nil
}

func (r *spaceRepository) SubscribeSub(userID uint64, spaceID uint64) error {
	s := models.Subscription{
		UserID:  userID,
		SpaceID: spaceID,
	}
	if err := r.db.Create(&s).Error; err != nil {
		return err
	}

	return nil
}

func (r *spaceRepository) UnsubscribeSub(userID, spaceID uint64) error {
	s := models.Subscription{
		UserID:  userID,
		SpaceID: spaceID,
	}
	if err := r.db.Model(&models.Subscription{}).Delete(&s, &s).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	return nil
}

func (r *spaceRepository) GetUserSubscriptions(userID uint64) ([]models.Space, error) {
	var spaces []models.Space
	if err := r.db.Joins("JOIN subscriptions ON subscriptions.space_id = spaces.id AND subscriptions.user_id=?", userID).Find(&spaces).Error; err != nil {
		return nil, err
	}

	return spaces, nil
}

func (r spaceRepository) IncreaseSpaceSubscribersCount(spaceId uint64, count int) error {
	return r.db.Model(&models.Space{}).Where("id=?", spaceId).Update("subscribers_count", gorm.Expr("subscribers_count + ?", count)).Error
}
