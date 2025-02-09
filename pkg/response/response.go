package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type jsonResponse struct{}

func NewJSONResponse() jsonResponse {
	return jsonResponse{}
}
func (j jsonResponse) Response(w http.ResponseWriter, status int, responseCode string, data map[string]any) {
	data["code"] = responseCode
	data["status"] = status

	w.Header().Add("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)

	slog.Info("request info",
		slog.Int("status", status),
		slog.Any("code", responseCode),
		slog.Any("data", data),
	)
}

func (j jsonResponse) ErrorResponse(w http.ResponseWriter, status int, responseCode string, data map[string]any, errs ...error) {
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
	j.Response(w, status, responseCode, data)
}

func (j jsonResponse) ServerErrorResponse(w http.ResponseWriter, err error) {
	slog.Error("SERVER ERROR", "error", err.Error())
	j.Response(w, http.StatusInternalServerError, "", map[string]any{"msg": "Internal server error"})
}

func (j jsonResponse) ServerOrUserErrorResponse(w http.ResponseWriter, serverErr error, userErrs []error, responseCode string) {
	if serverErr != nil {
		j.ServerErrorResponse(w, serverErr)
	} else if userErrs != nil {
		j.ErrorResponse(w, http.StatusBadRequest, responseCode, nil, userErrs...)
	} else {
		panic("func ServerOrUserErrorResponse: serverErr and userErrs are nil")
	}
}

func (j jsonResponse) InvalidJSONErrorResponse(w http.ResponseWriter, err error) {
	j.ErrorResponse(w, http.StatusBadRequest, "invalid_json", nil, errors.New("invalid json"))
}
