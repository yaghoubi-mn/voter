package response

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type JsonResponse struct{}

func NewJSONResponse() JsonResponse {
	return JsonResponse{}
}
func (j JsonResponse) Response(c *gin.Context, status int, responseCode string, data map[string]any) {
	data["code"] = responseCode
	data["status"] = status

	c.JSON(status, data)

	slog.Info("request info",
		slog.Int("status", status),
		slog.Any("code", responseCode),
		slog.Any("data", data),
	)
}

func (j JsonResponse) ErrorResponse(c *gin.Context, status int, responseCode string, data map[string]any, errs ...error) {
	if errs == nil {
		slog.Error("errs is required in ErrorResponse")
	}

	if data == nil {
		data = make(map[string]any)
	}

	temp := make(map[string]string)

	for _, err := range errs {
		splited := strings.Split(err.Error(), ": ")
		if len(splited) == 1 {
			temp["non_field"] = splited[0]
		} else {
			temp[splited[0]] = splited[1]
		}
	}

	data["errors"] = any(temp)
	j.Response(c, status, responseCode, data)
}

func (j JsonResponse) ServerErrorResponse(c *gin.Context, err error) {
	slog.Error("SERVER ERROR", "error", err.Error())
	j.Response(c, http.StatusInternalServerError, "", map[string]any{"msg": "Internal server error"})
}

func (j JsonResponse) ServerOrUserErrorResponse(c *gin.Context, serverErr error, userErrs []error, responseCode string) {
	if serverErr != nil {
		j.ServerErrorResponse(c, serverErr)
	} else if userErrs != nil {
		j.ErrorResponse(c, http.StatusBadRequest, responseCode, nil, userErrs...)
	} else {
		panic("func ServerOrUserErrorResponse: serverErr and userErrs are nil")
	}
}

func (j JsonResponse) InvalidJSONErrorResponse(c *gin.Context, err error) {
	j.ErrorResponse(c, http.StatusBadRequest, "invalid_json", nil, errors.New("invalid json"))
}
