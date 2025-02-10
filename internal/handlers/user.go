package handlers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/services"
	"github.com/yaghoubi-mn/voter/pkg/response"
)

type UserHandler interface {
	Login(c *gin.Context)
}

type userHandler struct {
	service  services.UserService
	response response.JsonResponse
}

func NewUserHandler(service services.UserService, response response.JsonResponse) UserHandler {
	return &userHandler{
		service:  service,
		response: response,
	}
}

func (h *userHandler) Login(c *gin.Context) {
	var loginInput dtos.LoginInput

	// decode body
	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&loginInput)
	if err != nil {
		h.response.InvalidJSONErrorResponse(c, err)
		return
	}

	responseDTO := h.service.Login(loginInput)
	if responseDTO.UserErrs != nil || responseDTO.ServerErr != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, 200, responseDTO.ResponseCode, responseDTO.Data)
}
