package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	docs "github.com/yaghoubi-mn/voter/docs"
	"github.com/yaghoubi-mn/voter/internal/cache"
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/database"
	"github.com/yaghoubi-mn/voter/internal/enums"
	"github.com/yaghoubi-mn/voter/internal/handlers"
	"github.com/yaghoubi-mn/voter/internal/middleware"
	"github.com/yaghoubi-mn/voter/internal/models"
	"github.com/yaghoubi-mn/voter/internal/permissions"
	"github.com/yaghoubi-mn/voter/internal/repositories"
	"github.com/yaghoubi-mn/voter/internal/routes"
	"github.com/yaghoubi-mn/voter/internal/services"
	"github.com/yaghoubi-mn/voter/pkg/jwt"
	"github.com/yaghoubi-mn/voter/pkg/response"
	"github.com/yaghoubi-mn/voter/pkg/utils"
	"gorm.io/gorm"
)

// @title Voter
// @version 1.0
// @BasePath /api/v1/
func main() {
	godotenv.Load()

	jwt.Init([]byte(os.Getenv("JWT_SECRET_KEY")))

	redisClient, err := cache.Setup()
	if err != nil {
		slog.Error("cache connection error", "error", err)
	}

	redisCache := cache.NewCache(redisClient, context.Background())

	db, err := database.Setup()
	if err != nil {
		slog.Error("cannot connect to database", "error", err.Error())
		return
	}

	// if len(os.Args) > 1 && os.Args[1] == "migrate" {
	migrate(db)
	// return
	// }

	addDefaultUsers(db)

	r := gin.Default()

	// setup swagger
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	fmt.Println("swagger URL: http://localhost:8080/swagger/index.html")

	validate := validator.New()
	response := response.NewJSONResponse()

	// middlewares
	authMiddleware := middleware.NewAuthMiddleware(response)

	// repositories
	userRepository := repositories.NewUserRepository(db)
	subRepository := repositories.NewSubRepository(db)
	postRepository := repositories.NewPostRepository(db)
	commentRepository := repositories.NewCommentRepository(db)
	commentVoteRepository := repositories.NewCommentVoteRepository(db)
	postVoteRepository := repositories.NewPostVoteRepository(db)

	// permissions
	subPermissions := permissions.NewSubPermission(&config.Settings{
		SubCreationPermission: enums.PermissionAll,
		SubClosePermission:    enums.PermissionAdmin,
		SubDeletePermission:   enums.PermissionAdmin,
	})

	// services
	userService := services.NewUserService(userRepository, validate)
	subService := services.NewSubService(subRepository, validate, subPermissions)
	postService := services.NewPostService(postRepository, postVoteRepository, validate, redisCache)
	commentService := services.NewCommentService(commentRepository, commentVoteRepository, postRepository, validate, redisCache)

	// handlers
	userHandler := handlers.NewUserHandler(userService, response)
	subHandler := handlers.NewSubHandler(subService, response)
	postHandler := handlers.NewPostHandler(postService, response)
	commentHandler := handlers.NewCommentHandler(commentService, response)

	routes.SetupRoutes(r, authMiddleware, userHandler, subHandler, postHandler, commentHandler)

	r.Run()
}

func migrate(db *gorm.DB) {
	db.AutoMigrate(
		models.User{},
		models.Post{},
		models.PostVote{},
		models.Comment{},
		models.CommentVote{},
	)
}

func addDefaultUsers(db *gorm.DB) {

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		slog.Error("error in getting users", "error", err)
	}

	if len(users) == 0 {
		// add default users
		users = []models.User{
			{
				ID:       1,
				Username: "admin",
				Password: "1234",
			},
			{
				ID:       2,
				Username: "test",
				Password: "test",
			},
		}
		for _, user := range users {

			var err error
			user.Salt, err = utils.GenerateRandomSalt()
			if err != nil {
				slog.Error("cannot generate salt", "error", err.Error())
				return
			}

			user.Password, err = utils.HashPasswordWithSalt(user.Password, user.Salt)
			if err != nil {
				slog.Error("cannot hash password", "error", err.Error())
				return
			}

			err = db.Create(&user).Error
			if err != nil {
				slog.Error("cannot insert user", "error", err.Error())
			}
		}

	}

}
