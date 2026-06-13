package services

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/internal/permissions"
	"github.com/yaghoubi-mn/voter/internal/repositories"
)

type SpaceService interface {
	Create(spaceInput dtos.SpaceCreateInput, user models.User) dtos.ResponseDTO
	Update(spaceInput dtos.SpaceEditInput, subId uint64, user models.User) dtos.ResponseDTO
	Delete(spaceId uint64, user models.User) dtos.ResponseDTO
	GetAll(sortBy enums.SortBy, page int) dtos.ResponseDTO
	GetByID(spaceId uint64) dtos.ResponseDTO
	Subscribe(spaceId uint64, user models.User) dtos.ResponseDTO
	Unsubscribe(spaceId uint64, user models.User) dtos.ResponseDTO
}

type spaceService struct {
	repo        repositories.SpaceRepository
	validate    *validator.Validate
	permissions permissions.SubPermission
}

func NewSubService(repo repositories.SpaceRepository, validate *validator.Validate, permissions permissions.SubPermission) SpaceService {
	return &spaceService{
		repo:        repo,
		validate:    validate,
		permissions: permissions,
	}
}

func (s *spaceService) Create(spaceInput dtos.SpaceCreateInput, user models.User) (responseDTO dtos.ResponseDTO) {

	if !s.permissions.HasCreationPermission(enums.Permissions(user.Role)) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to create space")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	errs := s.validate.Struct(spaceInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// save space to database
	space := spaceInput.GetSubModel(user.ID)

	if err := s.repo.Create(&space); err != nil {
		if err == custom_errors.DuplicateKey {
			responseDTO.UserErrs = []error{errors.New("username: this username is already taken")}
			responseDTO.ResponseCode = "invalid_username"
			responseDTO.Status = 400
		}
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Msg = "space created"
	spaceOutput := dtos.GetSubOutputFromSub(space)
	responseDTO.Data = spaceOutput
	return
}

func (s *spaceService) Update(spaceInput dtos.SpaceEditInput, spaceId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	errs := s.validate.Struct(spaceInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// get space from database
	var space models.Space
	space, err := s.repo.GetByID(spaceId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("space not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	// check user has access to sub
	if !s.permissions.HasEditPermission(user, space) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this space")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// update sub
	spaceInput.UpdateSub(&space)

	err = s.repo.Update(space)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Msg = "Done"
	return

}

func (s *spaceService) Delete(spaceId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	// get space from database
	space, err := s.repo.GetByID(spaceId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("space not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	// check user has permission
	if !s.permissions.HasDeletePermission(user, space) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this space")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// delete space
	err = s.repo.Delete(spaceId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Msg = "Done"
	return
}

func (s *spaceService) GetAll(sortBy enums.SortBy, page int) (responseDTO dtos.ResponseDTO) {

	var spaces []models.Space
	// get data from database
	spaces, err := s.repo.GetAll(sortBy, page)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	spacesOutput := make([]dtos.SpaceOutput, len(spaces))
	for i, sub := range spaces {
		spacesOutput[i] = dtos.GetSubOutputFromSub(sub)
	}

	responseDTO.Data = spacesOutput
	return
}

func (s *spaceService) GetByID(spaceId uint64) (responseDTO dtos.ResponseDTO) {
	var space models.Space

	// get the space from database
	space, err := s.repo.GetByID(spaceId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("space not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	spaceOutput := dtos.GetSubOutputFromSub(space)
	responseDTO.Data = spaceOutput
	return
}

func (s spaceService) Subscribe(spaceId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	if err := s.repo.SubscribeSub(user.ID, spaceId); err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23503") {
			responseDTO.UserErrs = []error{errors.New("spaceId: space id not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		} else if err == custom_errors.DuplicateKey {
			responseDTO.UserErrs = []error{errors.New("user is already subscribed to the space")}
			responseDTO.ResponseCode = "already_subscribed"
			responseDTO.Status = 400
			return
		}
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Msg = "Done"
	return
}

func (s spaceService) Unsubscribe(spaceId uint64, user models.User) (responseDTO dtos.ResponseDTO) {

	if err := s.repo.UnsubscribeSub(user.ID, spaceId); err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("user is not subscribed to the space")}
			responseDTO.ResponseCode = "not_subscribed"
			responseDTO.Status = 400
			return
		}
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Msg = "Done"
	return
}
