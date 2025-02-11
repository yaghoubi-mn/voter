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

type CommentService interface {
	Create(commentInput dtos.CommentInput, postId uint64, user models.User) dtos.ResponseDTO
	Update(commentInput dtos.CommentInput, commentId uint64, user models.User) dtos.ResponseDTO
	Delete(commentId uint64, user models.User) dtos.ResponseDTO
	GetAll(postId uint64, sortBy enums.SortBy, page int) dtos.ResponseDTO
	GetByID(commentId uint64) dtos.ResponseDTO

	Vote(commentId uint64, vote bool, user models.User) dtos.ResponseDTO
	DeleteVote(commentId uint64, user models.User) dtos.ResponseDTO
}

type commentService struct {
	repo     repositories.CommentRepository
	postRepo repositories.PostRepository
	voteRepo repositories.CommentVoteRepository
	validate *validator.Validate
	cache    cache.Cache
}

func NewCommentService(repo repositories.CommentRepository, voteRepo repositories.CommentVoteRepository, postRepo repositories.PostRepository, validate *validator.Validate, cache cache.Cache) CommentService {
	return &commentService{
		repo:     repo,
		voteRepo: voteRepo,
		postRepo: postRepo,
		validate: validate,
		cache:    cache,
	}
}

func (s *commentService) Create(commentInput dtos.CommentInput, postId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	errs := s.validate.Struct(commentInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// check post already exists
	_, err := s.postRepo.GetByID(postId)
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

	// check reply comment exists
	if commentInput.ReplyCommentID != 0 {
		_, err := s.repo.GetByID(commentInput.ReplyCommentID)
		if err != nil {
			if err == custom_errors.RecordNotFound {
				responseDTO.UserErrs = []error{errors.New("reply_comment_id: reply comment not found")}
				responseDTO.ResponseCode = "not_found"
				responseDTO.Status = 404
				return
			}

			responseDTO.ServerErr = err
			return
		}
	}

	// save comment to database
	var comment models.Comment
	comment.Content = commentInput.Content
	comment.AuthorID = user.ID
	comment.PostID = postId

	if err := s.repo.Create(comment); err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "comment created"
	return
}

func (s *commentService) Update(commentInput dtos.CommentInput, commentId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	errs := s.validate.Struct(commentInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// check comment blongs to user
	var comment models.Comment
	comment, err := s.repo.GetByID(commentId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if comment.AuthorID != user.ID {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this comment")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	if comment.IsDeleted {
		responseDTO.UserErrs = []error{errors.New("this comment has been deleted")}
		responseDTO.ResponseCode = "deleted_comment"
		responseDTO.Status = 400
		return
	}

	// update comment
	comment.Content = commentInput.Content
	comment.ModifiedAt = time.Now()
	err = s.repo.Update(comment)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("comment not found")}
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

func (s *commentService) Delete(commentId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// check comment blongs to user
	comment, err := s.repo.GetByID(commentId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("comment not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	if comment.AuthorID != user.ID {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this comment")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// check for the comment replies
	commentReplies, err := s.repo.GetAllReplies(commentId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if len(commentReplies) != 0 {

		// set comment as deleted
		comment.Content = "[deleted]"
		comment.IsDeleted = true
		err := s.repo.Update(comment)
		if err != nil {
			responseDTO.ServerErr = err
			return
		}

	} else {

		// delete comment
		err = s.repo.Delete(commentId)
		if err != nil {
			if err == custom_errors.RecordNotFound {
				responseDTO.UserErrs = []error{errors.New("comment not found")}
				responseDTO.ResponseCode = "not_found"
				responseDTO.Status = 404
				return
			}
			responseDTO.ServerErr = err
			return
		}

	}

	// clear cache
	err = s.cache.FlushDB()
	if err != nil {
		slog.Error("error in flushing database", "error", err)
	}

	responseDTO.Data["msg"] = "Done"
	return
}

func (s *commentService) GetAll(postId uint64, sortBy enums.SortBy, page int) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	cacheName := "comment_page_" + string(sortBy)

	var comments []models.Comment

	// get data from cache
	data, err := s.cache.Get(cacheName, postId)
	if err != nil {

		// get data from database
		comments, err = s.repo.GetAll(postId, sortBy, page)
		if err != nil {
			responseDTO.ServerErr = err
			return
		}

		// save data to cache
		data, err := json.Marshal(comments)
		if err != nil {
			slog.Error("cannot marshal data", "error", err)
		}
		err = s.cache.Set(cacheName, postId, string(data))
		if err != nil {
			slog.Error("cannot cache data", "error", err)
		}

	} else {
		json.NewDecoder(strings.NewReader(data)).Decode(&comments)
	}

	commentsOutput := make([]dtos.CommentOutput, len(comments))
	for i, comment := range comments {
		commentsOutput[i] = dtos.GetCommentOutputFromComment(comment)
	}

	responseDTO.Data["data"] = commentsOutput
	return
}

func (s *commentService) GetByID(commentId uint64) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	cacheName := "comment"

	var comment models.Comment
	// get data from cache
	data, err := s.cache.Get(cacheName, commentId)
	if err != nil {
		// get data from database
		comment, err = s.repo.GetByID(commentId)
		if err != nil {
			responseDTO.ServerErr = err
			return
		}

		// save data to cache
		data, err := json.Marshal(comment)
		if err != nil {
			slog.Error("cannot marshal data", "error", err)
		}
		err = s.cache.Set(cacheName, commentId, string(data))
		if err != nil {
			slog.Error("cannot cache data", "error", err)
		}

	} else {
		json.NewDecoder(strings.NewReader(data)).Decode(&comment)
	}
	commentOutput := dtos.GetCommentOutputFromComment(comment)
	responseDTO.Data["data"] = commentOutput
	return
}

func (s *commentService) Vote(commentId uint64, vote bool, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// check comment already exists
	_, err := s.repo.GetByID(commentId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("comment_id: comment not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}

		responseDTO.ServerErr = err
		return
	}

	// delete previous user vote if exists
	commentVote, err := s.voteRepo.Delete(commentId, user.ID)
	if err != nil && err != custom_errors.RecordNotFound {
		responseDTO.ServerErr = err
		return
	}

	previousVote := commentVote.Vote

	var newVote bool

	if err == custom_errors.RecordNotFound {
		newVote = true
	}

	commentVote = models.CommentVote{
		UserID:    user.ID,
		CommentID: commentId,
		Vote:      vote,
	}

	err = s.voteRepo.Create(commentVote)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	if newVote {
		if vote {
			// increase comment score
			err = s.repo.AddCommentScore(commentId, 1)

		} else {
			// decrease comment score
			err = s.repo.AddCommentScore(commentId, -1)
		}
	} else {

		if vote {
			if previousVote {
				// nothing
			} else {
				// comment score + 2
				err = s.repo.AddCommentScore(commentId, 2)
			}
		} else {
			if previousVote {
				// comment score - 2
				err = s.repo.AddCommentScore(commentId, -2)
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

func (s *commentService) DeleteVote(commentId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	// delete vote
	commentVote, err := s.voteRepo.Delete(commentId, user.ID)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("you didn't vote to this comment")}
			responseDTO.ResponseCode = "no_vote"
			responseDTO.Status = 400
			return
		}

		responseDTO.ServerErr = err
		return
	}

	if commentVote.Vote {
		err = s.repo.AddCommentScore(commentId, -1)
	} else {
		err = s.repo.AddCommentScore(commentId, 1)
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
