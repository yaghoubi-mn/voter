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
	GetBySpace(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
	UpVote(c *gin.Context)
	DownVote(c *gin.Context)
	DeleteVote(c *gin.Context)
	GetUserHomePosts(c *gin.Context)
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
// @Success 200 {object} response.SuccessResponse{data=dtos.PostOutput} "successfully created"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId/posts [post]
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
	subIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: sub id not found in url params"))
		return
	}
	subId, err := strconv.Atoi(subIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid post id"))
		return
	}

	responseDTO := h.service.Create(postInput, uint64(subId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// UpdatePost godoc
// @Description update a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param title body string true "post title"
// @Param content body string true "post content"
// @Success 200 {object} response.SuccessResponse "successfully updated"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403
// @Failure 500 {object} response.ErrorResponse "Internal server error"
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
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.Update(postInput, uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// DeletePost godoc
// @Description delete a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200 {object} response.SuccessResponse "successfully deleted"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
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
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.Delete(uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// GetAllPosts godoc
// @Description Get all posts by page. This URL is for home page.
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\" or \"score\""
// @Success 200 {object} response.SuccessResponse{data=[]dtos.PostOutput} "Successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403 {object} response.ErrorResponse "Access Denied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts [get]
func (h *postHandler) GetAll(c *gin.Context) {

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
	case "score":
		sortBy = enums.SortByScore
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

// GetAllPosts godoc
// @Description get all posts of a space by page
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\" or \"score\""
// @Success 200 {object} response.SuccessResponse{data=[]dtos.PostOutput} "Successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403 {object} response.ErrorResponse "Access Denied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /spaces/:spaceId/posts [get]
func (h *postHandler) GetBySpace(c *gin.Context) {

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
	case "score":
		sortBy = enums.SortByScore
	case "date":
		sortBy = enums.SortByDate
	case "":
		sortBy = enums.DefaultSort
	default:
		h.response.ErrorResponse(c, 400, "invalid_param", "", errors.New("sort_by: invalid sort_by value"))
		return
	}

	spaceIdString, ok := c.Params.Get("spaceId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "space_id_not_found_in_url", "", errors.New("space_id: space id not found in url params"))
		return
	}
	spaceId, err := strconv.Atoi(spaceIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_space_id", "", errors.New("space_id: invalid post id"))
		return
	}

	// call service
	responseDTO := h.service.GetBySpace(sortBy, page, uint64(spaceId))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// GetPost godoc
// @Description get a post
// @Tags posts
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=dtos.PostOutput} "successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts/:postId [get]
func (h *postHandler) GetByID(c *gin.Context) {

	// get postId from url
	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	responseDTO := h.service.GetByID(uint64(postId))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}

// UpvotePost godoc
// @Description upvote to a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200 {object} response.SuccessResponse "successfully saved"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts/:postId/upvote [post]
func (h *postHandler) UpVote(c *gin.Context) {

	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.Vote(uint64(postId), true, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)

}

// DownVotePost godoc
// @Description downvote to a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200 {object} response.SuccessResponse "successfully saved"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts/:postId/downvote [post]
func (h *postHandler) DownVote(c *gin.Context) {

	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.Vote(uint64(postId), false, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)

}

// DeleteVotePost godoc
// @Description delete vote of a post
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Success 200 {object} response.SuccessResponse "successfully saved"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts/:postId/votes [delete]
func (h *postHandler) DeleteVote(c *gin.Context) {

	postIdString, ok := c.Params.Get("postId")
	if !ok {
		h.response.ErrorResponse(c, http.StatusBadRequest, "post_id_not_found_in_url", "", errors.New("post_id: post id not found in url params"))
		return
	}
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		h.response.ErrorResponse(c, http.StatusBadRequest, "invalid_post_id", "", errors.New("post_id: invalid post id"))
		return
	}

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

	responseDTO := h.service.DeleteVote(uint64(postId), user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)

}

// GetUserHomePosts godoc
// @Description Get user home posts. can be sort by "trending", "score", "date".
// @Tags posts
// @Accept json
// @Produce json
// @Param Authorization header string true "authorization token (value: Bearer <jwt-token>)"
// @Param page query integer true "page number"
// @Param sort_by query string true "\"date\" or \"score\" or \"trending\""
// @Success 200 {object} response.SuccessResponse{data=[]dtos.PostOutput} "Successfully fetched"
// @Failure 400 {object} response.ErrorResponse "falied"
// @Failure 403 {object} response.ErrorResponse "Access Denied"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /posts/home [get]
func (h *postHandler) GetUserHomePosts(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		h.response.ServerErrorResponse(c, errors.New("user not found in context"))
		return
	}

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
	case "score":
		sortBy = enums.SortByScore
	case "date":
		sortBy = enums.SortByDate
	case "trending":
		sortBy = enums.Trending
	case "":
		sortBy = enums.SortByDate
	default:
		h.response.ErrorResponse(c, 400, "invalid_param", "", errors.New("sort_by: invalid sort_by value"))
		return
	}

	// call service
	responseDTO := h.service.GetUserHomePosts(sortBy, page, user.(models.User))
	if responseDTO.ServerErr != nil || responseDTO.UserErrs != nil {
		h.response.ServerOrUserErrorResponse(c, responseDTO.Status, responseDTO.Msg, responseDTO.ServerErr, responseDTO.UserErrs, responseDTO.ResponseCode)
		return
	}

	h.response.Response(c, http.StatusOK, responseDTO.ResponseCode, responseDTO.Msg, responseDTO.Data, nil)
}
