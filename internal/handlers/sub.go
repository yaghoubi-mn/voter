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

type SubHandler interface {
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
}

type subHandler struct {
	service  services.SubService
	response response.JsonResponse
}

func NewSubHandler(service services.SubService, response response.JsonResponse) SubHandler {
	return &subHandler{
		service:  service,
		response: response,
	}
}

// CreatePost godoc
// @Description create a space
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "space title"
// @Param description body string true "space description"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /spaces [post]
func (h *subHandler) Create(c *gin.Context) {
	var subInput dtos.SubInput

	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	decoder.DisallowUnknownFields()
	decoder.Decode(&subInput)

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.Create(subInput, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

// UpdatePost godoc
// @Description update a space title or description
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "space title"
// @Param description body string true "space description"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /spaces/:spaceId [put]
func (h *subHandler) Update(c *gin.Context) {
	var subInput dtos.SubInput

	decoder := json.NewDecoder(c.Request.Body)
	defer c.Request.Body.Close()
	decoder.DisallowUnknownFields()
	decoder.Decode(&subInput)

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get subId from url
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", nil, errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", nil, errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.Update(subInput, uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

// DeletePost godoc
// @Description delete a space
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /spaces/:spaceId [delete]
func (h *subHandler) Delete(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get spaceId from url
	spaceIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", nil, errors.New("space_id: space id not found in url params"))
		return
	}
	spaceId, err := strconv.Atoi(spaceIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", nil, errors.New("sub_id: invalid space id"))
		return
	}

	responseDTO := h.service.Delete(uint64(spaceId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

// GetAllPosts godoc
// @Description get all spaces by page
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /spaces [get]
func (h *subHandler) GetAll(c *gin.Context) {

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

	sortByString, ok := c.GetQuery("sort_by")

	var sortBy enums.SortBy
	if !ok {
		sortBy = enums.DefaultSort
	}

	switch sortByString {
	case "date":
		sortBy = enums.SortByDate
	case "":
		sortBy = enums.DefaultSort
	default:
		h.response.ErrorResponse(c, 400, "invalid_param", nil, errors.New("sort_by: invalid sort_by value"))
		return
	}

	// call service
	responseDTO := h.service.GetAll(sortBy, page)
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

// GetPost godoc
// @Description get a space by ID
// @Tags spaces
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 500
// @Router /spaces/:spaceId [get]
func (h *subHandler) GetByID(c *gin.Context) {

	// get subId from url
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", nil, errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", nil, errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.GetByID(uint64(subId))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}
