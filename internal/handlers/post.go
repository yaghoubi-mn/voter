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
	UpVote(c *gin.Context)
	DownVote(c *gin.Context)
	DeleteVote(c *gin.Context)
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

// CreatePost godoc
// @Description create a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "post title"
// @Param content body string true "post content"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /subs/:subId/posts [post]
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

	// get subId from url
	subIdString, ok := c.Params.Get("subId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "sub_id_not_found_in_url", nil, errors.New("sub_id: sub id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_sub_id", nil, errors.New("sub_id: invalid post id"))
		return
	}

	responseDTO := h.service.Create(postInput, uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)
}

// UpdatePost godoc
// @Description update a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "post title"
// @Param content body string true "post content"
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /posts/:postId [put]
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

// DeletePost godoc
// @Description delete a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /posts/:postId [delete]
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

// GetAllPosts godoc
// @Description get all posts by page
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\" or \"score\""
// @Success 200
// @Failure 400
// @Failure 403
// @Failure 500
// @Router /posts/ [get]
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

	sortByString, ok := c.GetQuery("sort_by")

	var sortBy enums.SortBy
	if !ok {
		sortBy = enums.DefaultSort
	}

	switch sortByString {
	case "score":
		sortBy = enums.SortByScore
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
// @Description get a post
// @Tags posts
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 500
// @Router /posts/:postId [get]
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

// UpvotePost godoc
// @Description upvote to a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /posts/:postId/upvote [post]
func (h *postHandler) UpVote(c *gin.Context) {

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

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.Vote(uint64(postId), true, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)

}

// DownVotePost godoc
// @Description downvote to a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /posts/:postId/downvote [post]
func (h *postHandler) DownVote(c *gin.Context) {

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

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.Vote(uint64(postId), false, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)

}

// DeleteVotePost godoc
// @Description delete vote of a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /posts/:postId/votes [delete]
func (h *postHandler) DeleteVote(c *gin.Context) {

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

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.DeleteVote(uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Data)

}
