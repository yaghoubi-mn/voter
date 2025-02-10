package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/yaghoubi-mn/voter/internal/handlers"
	"github.com/yaghoubi-mn/voter/internal/middleware"
)

func SetupRoutes(r *gin.Engine, authMiddleware middleware.AuthMiddleware, userHandlers handlers.UserHandler, postHandlers handlers.PostHandler, commentHandler handlers.CommentHandler) {

	// middlewares
	// r.Use(middleware.Auth())

	v1 := r.Group("/api/v1")

	// users
	v1.POST("/users/login", userHandlers.Login)

	authV1 := v1.Group("/")
	authV1.Use(authMiddleware.Auth())

	// posts
	authV1.POST("/posts", postHandlers.Create)
	authV1.PUT("/posts/:postId", postHandlers.Update)
	authV1.DELETE("/posts/:postId", postHandlers.Delete)
	v1.GET("/posts", postHandlers.GetAll)
	v1.GET("/posts/:postId", postHandlers.GetByID)

	// comments
	authV1.POST("posts/:postId/comments", commentHandler.Create)
	authV1.PUT("/comments/:commentId", commentHandler.Update)
	authV1.DELETE("/comments/:commentId", commentHandler.Delete)
	v1.GET("/posts/:postId/comments", commentHandler.GetAll)

}
