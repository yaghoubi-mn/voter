package services

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/yaghoubi-mn/voter/internal/custom_errors"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/internal/permissions"
	"github.com/yaghoubi-mn/voter/internal/repositories"
)

type SubService interface {
	Create(subInput dtos.SubInput, user models.User) dtos.ResponseDTO
	Update(subInput dtos.SubInput, subId uint64, user models.User) dtos.ResponseDTO
	Delete(subId uint64, user models.User) dtos.ResponseDTO
	GetAll(sortBy enums.SortBy, page int) dtos.ResponseDTO
	GetByID(subId uint64) dtos.ResponseDTO
}

type subService struct {
	repo        repositories.SubRepository
	validate    *validator.Validate
	permissions permissions.SubPermission
}

func NewSubService(repo repositories.SubRepository, validate *validator.Validate, permissions permissions.SubPermission) SubService {
	return &subService{
		repo:        repo,
		validate:    validate,
		permissions: permissions,
	}
}

func (s *subService) Create(subInput dtos.SubInput, user models.User) (responseDTO dtos.ResponseDTO) {

	responseDTO.Data = make(map[string]any)

	if s.permissions.HasCreationPermission(enums.Permissions(user.Role)) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to create sub")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	errs := s.validate.Struct(subInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// save sub to database
	sub := subInput.GetSubModel(user.ID)

	if err := s.repo.Create(sub); err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "sub created"
	return
}

func (s *subService) Update(subInput dtos.SubInput, subId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	errs := s.validate.Struct(subInput)
	if errs != nil {
		errList := make([]error, 0, 2)
		for _, err := range errs.(validator.ValidationErrors) {
			errList = append(errList, errors.New(err.StructField()+": "+err.Error()))
		}

		responseDTO.ResponseCode = "invalid_field"
		responseDTO.UserErrs = errList
		return
	}

	// get sub from database
	var sub models.Sub
	sub, err := s.repo.GetByID(subId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("sub not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	// check user has access to sub
	if s.permissions.HasEditPermission(user, sub) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this sub")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// update sub
	subInput.UpdateSub(&sub)

	err = s.repo.Update(sub)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "Done"
	return

}

func (s *subService) Delete(subId uint64, user models.User) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	// get sub from database
	sub, err := s.repo.GetByID(subId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("sub not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	// check user has permission
	if s.permissions.HasDeletePermission(user, sub) {
		responseDTO.UserErrs = []error{errors.New("you havn't access to this sub")}
		responseDTO.ResponseCode = "access_denied"
		responseDTO.Status = 403
		return
	}

	// delete sub
	err = s.repo.Delete(subId)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	responseDTO.Data["msg"] = "Done"
	return
}

func (s *subService) GetAll(sortBy enums.SortBy, page int) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	var subs []models.Sub
	// get data from database
	subs, err := s.repo.GetAll(sortBy, page)
	if err != nil {
		responseDTO.ServerErr = err
		return
	}

	subsOutput := make([]dtos.SubOutput, len(subs))
	for i, sub := range subs {
		subsOutput[i] = dtos.GetSubOutputFromSub(sub)
	}

	responseDTO.Data["data"] = subsOutput
	return
}

func (s *subService) GetByID(subId uint64) (responseDTO dtos.ResponseDTO) {
	responseDTO.Data = make(map[string]any)

	var sub models.Sub

	// get the sub from database
	sub, err := s.repo.GetByID(subId)
	if err != nil {
		if err == custom_errors.RecordNotFound {
			responseDTO.UserErrs = []error{errors.New("sub not found")}
			responseDTO.ResponseCode = "not_found"
			responseDTO.Status = 404
			return
		}
		responseDTO.ServerErr = err
		return
	}

	subOutput := dtos.GetSubOutputFromSub(sub)
	responseDTO.Data["data"] = subOutput
	return
}
