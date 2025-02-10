package repositories

import (
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post models.Post) error
	Update(post models.Post) error
	Delete(postId uint64) error
	GetByID(postId uint64) (models.Post, error)
	GetAll(sortBy enums.SortBy, page int) ([]models.Post, error)
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

func (r *postRepository) Create(post models.Post) error {
	return r.db.Create(&post).Error
}

func (r *postRepository) Update(post models.Post) error {
	return r.db.Updates(&post).Error

}

func (r *postRepository) Delete(postId uint64) error {
	return r.db.Delete(models.Post{ID: postId}).Error
}

func (r *postRepository) GetByID(postId uint64) (models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Author").First(&post, &models.Post{ID: postId}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return post, custom_errors.RecordNotFound
		}

		return post, err
	}

	return post, nil
}

func (r *postRepository) GetAll(sortBy enums.SortBy, page int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("Author").Order(sortBy).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&posts).Error

	return posts, err
}
