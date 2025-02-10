package services

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/internal/repositories"
)

type PostService interface {
	Create(postInput dtos.PostInput, user models.User) dtos.ResponseDTO
	Update(postInput dtos.PostInput, postId uint64, user models.User) dtos.ResponseDTO
	Delete(postId uint64, user models.User) dtos.ResponseDTO
	GetAll(sortBy enums.SortBy, page int) dtos.ResponseDTO
	GetByID(postId uint64) dtos.ResponseDTO
}

type postService struct {
	repo     repositories.PostRepository
	validate *validator.Validate
}

func NewPostService(repo repositories.PostRepository, validate *validator.Validate) PostService {
	return &postService{
		repo:     repo,
		validate: validate,
	}
}

func (s *postService) Create(postInput dtos.PostInput, user models.User) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	errs := s.validate.Struct(postInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// save post to database
	var post models.Post
	post.Title = postInput.Title
	post.Content = postInput.Content
	post.AuthorID = user.ID

	if err := s.repo.Create(post); err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "post created"
	return
}

func (s *postService) Update(postInput dtos.PostInput, postId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	errs := s.validate.Struct(postInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// check post blongs to user
	var post models.Post
	post, err := s.repo.GetByID(postId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if post.AuthorID != user.ID {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this post")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// update post
	post.Title = postInput.Title
	post.Content = postInput.Content
	post.ModifiedAt = time.Now()
	err = s.repo.Update(post)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("post not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "Done"
	return

}

func (s *postService) Delete(postId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// check post blongs to user
	post, err := s.repo.GetByID(postId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if post.AuthorID != user.ID {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this post")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// delete post
	err = s.repo.Delete(postId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("post not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "Done"
	return
}

func (s *postService) GetAll(sortBy enums.SortBy, page int) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// get data from database
	posts, err := s.repo.GetAll(sortBy, page)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	postsOutput := make([]dtos.PostOutput, len(posts))
	for i, post := range posts {
		postsOutput[i] = dtos.GetPostOutputFromPost(post)
	}

	responseDTO.Data["data"] = postsOutput
	return
}

func (s *postService) GetByID(postId uint64) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	post, err := s.repo.GetByID(postId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	postOutput := dtos.GetPostOutputFromPost(post)
	responseDTO.Data["data"] = postOutput
	return
}
