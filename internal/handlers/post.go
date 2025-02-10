package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yaghoubi-mn/voter/internal/dtos"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/internal/services"
	"github.com/yaghoubi-mn/voter/pkg/response"
)

type PostHandler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
}

type postHandler struct {
	service  services.PostService
	response response.JsonResponse
}

func NewPostHandler(service services.PostService, response response.JsonResponse) PostHandler {
	return &postHandler{
		service:  service,
		response: response,
	}
}

func (h *postHandler) Create(c *gin.Context) {
	var postInput dtos.PostInput

	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	decoder.DisallowUnknownFields()
	decoder.Decode(&postInput)

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}
	responseDTO := h.service.Create(postInput, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

func (h *postHandler) Update(c *gin.Context) {
	var postInput dtos.PostInput

	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	decoder.DisallowUnknownFields()
	decoder.Decode(&postInput)

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get postId from url
	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", nil, errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", nil, errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.Update(postInput, uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

func (h *postHandler) Delete(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get postId from url
	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", nil, errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", nil, errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.Delete(uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

func (h *postHandler) GetAll(c *gin.Context) {

	// get query params from url
	pageString := c.Query("page")
	if pageString == "" {
		h.response.ErrorResponse(c, http.StatusBadRequest, "page_not_found_in_url", nil, errors.New("page: page is required. ex: /?page=1"))
		return
	}
	page, err := strconv.Atoi(pageString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_page", nil, errors.New("page: invalid page. page must be integer"))
		return
	}

	orderByString := c.Query("order_by")

	var orderBy enums.SortBy
	switch orderByString {
	case "score":
		orderBy = enums.SortByScore
	case "date":
		orderBy = enums.SortByDate
	default:
		orderBy = enums.DefaultSort
	}

	// call service
	responseDTO := h.service.GetAll(orderBy, page)
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

func (h *postHandler) GetByID(c *gin.Context) {

	// get postId from url
	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", nil, errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", nil, errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.GetByID(uint64(postId))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}
