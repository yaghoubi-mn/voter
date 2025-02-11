package services

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/yaghoubi-mn/voter/internal/cache"
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
	Vote(postId uint64, vote bool, user models.User) dtos.ResponseDTO
	DeleteVote(postId uint64, user models.User) dtos.ResponseDTO
}

type postService struct {
	repo     repositories.PostRepository
	voteRepo repositories.PostVoteRepository
	validate *validator.Validate
	cache    cache.Cache
}

func NewPostService(repo repositories.PostRepository, voteRepo repositories.PostVoteRepository, validate *validator.Validate, cache cache.Cache) PostService {
	return &postService{
		repo:     repo,
		voteRepo: voteRepo,
		validate: validate,
		cache:    cache,
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

	// clear cache
	err = s.cache.FlushDB()
	if err != nil {
		slog.Error("error in flushing database", "error", err)
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

	// clear cache
	err = s.cache.FlushDB()
	if err != nil {
		slog.Error("error in flushing database", "error", err)
	}

	responseDTO.Data["msg"] = "Done"
	return
}

func (s *postService) GetAll(sortBy enums.SortBy, page int) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	cacheName := "post_page_" + string(sortBy)

	var posts []models.Post
	// get data from database
	data, err := s.cache.Get(cacheName, uint64(page))
	if err != nil {

		// get data from database
		posts, err = s.repo.GetAll(sortBy, page)
		if err != nil {
			responseDTO.ServerErr = err
			return
		}

		// save posts to cache
		data, err := json.Marshal(posts)
		if err != nil {

			slog.Error("cannot marshal data", "error", err)

		} else {

			err = s.cache.Set(cacheName, uint64(page), string(data))
			if err != nil {
				slog.Error("cannot cache data", "error", err)
			}
		}

	} else {
		json.NewDecoder(strings.NewReader(data)).Decode(&posts)
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

	cacheName := "post"

	var post models.Post

	data, err := s.cache.Get(cacheName, postId)
	if err != nil {

		// get data from database
		post, err = s.repo.GetByID(postId)
		if err != nil {
			responseDTO.ServerErr = err
			return
		}

		data, err := json.Marshal(post)
		if err != nil {
			slog.Error("cannot marshal post", "error", err)
		} else {

			err = s.cache.Set(cacheName, postId, string(data))
			if err != nil {
				slog.Error("cannot cache data", "error", err)
			}

		}

	} else {
		json.NewDecoder(strings.NewReader(data)).Decode(&post)
	}

	postOutput := dtos.GetPostOutputFromPost(post)
	responseDTO.Data["data"] = postOutput
	return
}

func (s *postService) Vote(postId uint64, vote bool, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// check post already exists
	_, err := s.repo.GetByID(postId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("post_id: post not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}

		responseDTO.ServerErr = err
		return
	}

	// delete previous user vote if exists
	postVote, err := s.voteRepo.Delete(postId, user.ID)
	if err != nil && err != custom_errors.RecordNotFound {
		responseDTO.ServerErr = err
		return
	}

	previousVote := postVote.Vote

	var newVote bool

	if err == custom_errors.RecordNotFound {
		newVote = true
	}

	postVote = models.PostVote{
		UserID: user.ID,
		PostID: postId,
		Vote:   vote,
	}

	err = s.voteRepo.Create(postVote)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if newVote {
		if vote {
			// increase post score
			err = s.repo.AddPostScore(postId, 1)

		} else {
			// decrease post score
			err = s.repo.AddPostScore(postId, -1)
		}
	} else {

		if vote {
			if previousVote {
				// nothing
			} else {
				// post score + 2
				err = s.repo.AddPostScore(postId, 2)
			}
		} else {
			if previousVote {
				// post score - 2
				err = s.repo.AddPostScore(postId, -2)
			} else {
				// nothing
			}
		}
	}
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	// clear cache
	err = s.cache.FlushDB()
	if err != nil {
		slog.Error("error in flushing database", "error", err)
	}

	responseDTO.Data["msg"] = "Done"
	return

}

func (s *postService) DeleteVote(postId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	// delete vote
	postVote, err := s.voteRepo.Delete(postId, user.ID)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("you didn't vote to this post")}
			responseDTO.ResponseCode = "no_vote"
			responseDTO.Status = 400
			return
		}

		responseDTO.ServerErr = err
		return
	}

	if postVote.Vote {
		err = s.repo.AddPostScore(postId, -1)
	} else {
		err = s.repo.AddPostScore(postId, 1)
	}
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	// clear cache
	err = s.cache.FlushDB()
	if err != nil {
		slog.Error("error in flushing database", "error", err)
	}

	responseDTO.Data["msg"] = "Done"
	return
}
