package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/pkg/jwt"
	"github.com/yaghoubi-mn/voter/pkg/response"
)

type AuthMiddleware struct {
	response response.JsonResponse
}

func NewAuthMiddleware(response response.JsonResponse) AuthMiddleware {
	return AuthMiddleware{
		response: response,
	}
}

func (m *AuthMiddleware) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		access := ctx.Request.Header.Get("Authorization")
		if access == "" {
			m.response.ErrorResponse(ctx, http.StatusUnauthorized, "", nil, errors.New("authentication is required"))
			ctx.Abort()
			return
		}

		if strings.Index(access, "Bearer ") != 0 || len(access) < 8 {
			m.response.ErrorResponse(ctx, http.StatusBadRequest, "invalid_header", nil, errors.New("invalid authorization header format"))
			ctx.Abort()
			return
		}

		access = access[7:]

		var user models.User
		var err error
		user.ID, user.Username, err = jwt.GetUserFromAccess(access)
		if err != nil {
			slog.Info("jwt erorr", "error", err)
			m.response.ErrorResponse(ctx, http.StatusBadRequest, "invalid_token", nil, errors.New("authorization: invalid token"))
			ctx.Abort()
			return
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}
