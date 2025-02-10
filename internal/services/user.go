package services

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/repositories"
	"github.com/yaghoubi-mn/voter/pkg/jwt"
	"github.com/yaghoubi-mn/voter/pkg/utils"
)

type UserService interface {
	Login(loginInput dtos.LoginInput) dtos.ResponseDTO
}

type userService struct {
	repo     repositories.UserRepository
	validate *validator.Validate
}

func NewUserService(userRepository repositories.UserRepository, validate *validator.Validate) UserService {
	return &userService{
		repo:     userRepository,
		validate: validate,
	}
}

func (s *userService) Login(loginInput dtos.LoginInput) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	// validate inputs
	errs := s.validate.Struct(loginInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// get user from database
	user, err := s.repo.GetByUsername(loginInput.Username)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{err}
			responseDTO.ResponseCode = "user_not_found"
			return
		}

		responseDTO.ServerErr = err
		return

	}

	// check password
	err = utils.CompareHashAndPassword(user.Password, loginInput.Password, user.Salt)
	if err != nil {
		slog.Error("err", "errr", err.Error())
		responseDTO.UserErrs = []error{errors.New("wrong password")}
		responseDTO.ResponseCode = "wrong_password"
		return
	}

	// generate jwt
	tokens, err := jwt.CreateRefreshAndAccessFromUserWithMap(config.JWTRefreshExpireTime, config.JWTAccessExpireTime, user.ID, user.Username)
	responseDTO.Data["tokens"] = tokens
	if err != nil {
		responseDTO.UserErrs = []error{err}
	}
	return

}
