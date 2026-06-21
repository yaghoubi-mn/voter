package repositories

import (
	"fmt"
	"time"

	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"gorm.io/gorm"
)

type PostRepository interface {
	Create(post *models.Post) error
	Update(post models.Post) error
	Delete(postId uint64) error
	GetByID(postId uint64) (models.Post, error)
	GetAll(sortBy enums.SortBy, page int) ([]models.Post, error)
	GetBySpace(sortBy enums.SortBy, page int, spaceId uint64) ([]models.Post, error)
	GetUserHomePosts(sortBy enums.SortBy, page int, userID uint64) ([]models.Post, error)

	AddPostScore(postId uint64, number int) error
	IncreasePostCommentsCount(postId uint64, count int) error
	IncreasePostViews(postId uint64, count int) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

func (r *postRepository) Create(post *models.Post) error {

	if err := r.db.Create(post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return custom_errors.RecordNotFound
		}
		return err
	}
	return nil
}

func (r *postRepository) Update(post models.Post) error {
	return r.db.Updates(&post).Error

}

func (r *postRepository) Delete(postId uint64) error {
	return r.db.Delete(models.Post{ID: postId}).Error
}

func (r *postRepository) GetByID(postId uint64) (models.Post, error) {
	var post models.Post
	if err := r.db.Preload("Author").Preload("Space").First(&post, &models.Post{ID: postId}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return post, custom_errors.RecordNotFound
		}

		return post, err
	}

	return post, nil
}

func (r *postRepository) GetAll(sortBy enums.SortBy, page int) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("Author").Preload("Space").Order(string(sortBy)).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&posts).Error

	return posts, err
}

func (r *postRepository) GetBySpace(sortBy enums.SortBy, page int, spaceId uint64) ([]models.Post, error) {
	var posts []models.Post
	err := r.db.Preload("Author").Order(string(sortBy)).Where(models.Post{SpaceID: spaceId}).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&posts).Error

	return posts, err
}

func (r *postRepository) AddPostScore(postId uint64, number int) error {
	var expr string
	if number >= 0 {
		expr = fmt.Sprintf("score + %v", number)
	} else {
		expr = fmt.Sprintf("score %v", number)
	}

	if err := r.db.Model(&models.Post{}).Where("id=?", postId).Update("score", gorm.Expr(expr)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return custom_errors.RecordNotFound
		}

		return err
	}

	return nil
}

func (r *postRepository) GetUserHomePosts(sortBy enums.SortBy, page int, userID uint64) ([]models.Post, error) {
	var posts []models.Post
	var err error
	var order any = string(sortBy)
	if sortBy == enums.Trending {
		order = gorm.Expr("(posts.score + posts.comments_count) / POWER(TIMESTAMPDIFF(HOUR, created_at, NOW())+ 2, %f) DESC", config.PostTrendingGravity)
	}
	err = r.db.Preload("Author").Preload("Space").Where("posts.created_at > ?", time.Now().AddDate(0, 0, -7)).
		Joins("JOIN subscriptions ON subscriptions.space_id = posts.space_id AND subscriptions.user_id = ?", userID).
		Order(order).Offset((page - 1) * config.PageLimit).Limit(config.PageLimit).Find(&posts).Error

	return posts, err
}

func (r postRepository) IncreasePostCommentsCount(postId uint64, count int) error {
	return r.db.Model(&models.Post{}).Where("id=?", postId).Update("comments_count", gorm.Expr("comments_count + ?", count)).Error
}

func (r postRepository) IncreasePostViews(postId uint64, count int) error {
	return r.db.Model(&models.Post{}).Where("id=?", postId).Update("views", gorm.Expr("views + ?", count)).Error
}
