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
	Subscribe(c *gin.Context)
	Unsubscribe(c *gin.Context)
}

type spaceHandler struct {
	service  services.SpaceService
	response response.JsonResponse
}

func NewSubHandler(service services.SpaceService, response response.JsonResponse) SubHandler {
	return &spaceHandler{
		service:  service,
		response: response,
	}
}

// CreateSpace godoc
// @Description create a space
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "space title"
// @Param description body string true "space description"
// @Success 200 {object} response.SuccessResponse{data=dtos.SpaceOutput} "successfully created"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces [post]
func (h *spaceHandler) Create(c *gin.Context) {
	var subInput dtos.SpaceCreateInput

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
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// UpdateSpace godoc
// @Description update a space title or description
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "space title"
// @Param description body string true "space description"
// @Success 200 {object} response.SuccessResponse "successfully updated"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId [put]
func (h *spaceHandler) Update(c *gin.Context) {
	var subInput dtos.SpaceEditInput

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
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.Update(subInput, uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// DeleteSpace godoc
// @Description delete a space
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200 {object} response.SuccessResponse "successfully deleted"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId [delete]
func (h *spaceHandler) Delete(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get spaceId from url
	spaceIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	spaceId, err := strconv.Atoi(spaceIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("sub_id: invalid space id"))
		return
	}

	responseDTO := h.service.Delete(uint64(spaceId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// GetAllSpaces godoc
// @Description Get all spaces by page. Page paramenters in URL is required!
// @Tags spaces
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\"
// @Success 200 {object} response.SuccessResponse{data=[]dtos.SpaceOutput} "successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces [get]
func (h *spaceHandler) GetAll(c *gin.Context) {

	// get query params from url
	pageString := c.Query("page")
	if pageString == "" {
		h.response.ErrorResponse(c, http.StatusBadRequest, "page_not_found_in_url", "", errors.New("page: page is required. ex: /?page=1"))
		return
	}
	page, err := strconv.Atoi(pageString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_page", "", errors.New("page: invalid page. page must be integer"))
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
		h.response.ErrorResponse(c, 400, "invalid_param", "", errors.New("sort_by: invalid sort_by value"))
		return
	}

	// call service
	responseDTO := h.service.GetAll(sortBy, page)
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// GetSpace godoc
// @Description get a space by ID
// @Tags spaces
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=dtos.SpaceOutput} "successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId [get]
func (h *spaceHandler) GetByID(c *gin.Context) {

	// get subId from url
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.GetByID(uint64(subId))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// Subscribe godoc
// @Description Subscribe a space
// @Tags spaces
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse "successfully done"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId/subscribe [post]
func (h *spaceHandler) Subscribe(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get subId from url
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.Subscribe(uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// UnsubscribeSpace godoc
// @Description Unsubscribe a space
// @Tags spaces
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse "successfully done"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId/unsubscribe [post]
func (h *spaceHandler) Unsubscribe(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	// get subId from url
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid space id"))
		return
	}

	responseDTO := h.service.Unsubscribe(uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}
