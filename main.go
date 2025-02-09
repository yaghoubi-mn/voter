package main

import (
	"log/slog"

	"github.com/yaghoubi-mn/voter/internal/database"
	"github.com/yaghoubi-mn/voter/internal/handlers"
	"github.com/yaghoubi-mn/voter/internal/repositories"
	"github.com/yaghoubi-mn/voter/internal/routes"
	"github.com/yaghoubi-mn/voter/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	db, err := database.Setup()
	if err != nil {
		slog.Error("cannot connect to database", "error", err.Error())
		return
	}

	userRepository := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepository)
	handlers := handlers.NewUserHandler(userService)
	routes.SetupRoutes(r, handlers)

	r.Run()
}
