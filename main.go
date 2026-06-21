package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	docs "github.com/yaghoubi-mn/voter/docs"
	"github.com/yaghoubi-mn/voter/internal/cache"
	"github.com/yaghoubi-mn/voter/internal/config"
	"github.com/yaghoubi-mn/voter/internal/database"
	"github.com/yaghoubi-mn/voter/internal/dtos"
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

// @title Veteria (Previously Voter)
// @version 1.0
// @BasePath /api/v1/
func main() {
	godotenv.Load()

	jwt.Init([]byte(os.Getenv("JWT_SECRET_KEY")))

	// Setup redis cache
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

	// setup cors middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})) // change this on production

	// setup swagger
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	fmt.Println("swagger URL: http://localhost:8080/swagger/index.html")

	validate := setupValidator()
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

	if len(os.Args) > 1 && os.Args[1] == "test-data" {
		addTestData(userRepository, userService, subService, postService, commentService)
		return
	}

	r.Run()
}

func migrate(db *gorm.DB) {
	db.AutoMigrate(
		models.Space{},
		models.User{},
		models.Post{},
		models.PostVote{},
		models.Comment{},
		models.CommentVote{},
		models.Subscription{},
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

func setupValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterValidation("username", utils.ValidateUsername)
	return validate
}

func addTestData(userRepository repositories.UserRepository, userService services.UserService, spaceService services.SpaceService, postService services.PostService, commentService services.CommentService) {
	// create two users
	userService.Register(dtos.RegisterInput{Username: "user1", Password: "12345678"})
	userService.Register(dtos.RegisterInput{Username: "user2", Password: "12345678"})

	// get users
	user1, _ := userRepository.GetByUsername("user1")
	user2, _ := userRepository.GetByUsername("user2")

	// create some spaces
	space1 := spaceService.Create(dtos.SpaceCreateInput{Title: "First space", Username: "first_space", Description: "This is a discription"}, user1).Data.(dtos.SpaceOutput)
	space2 := spaceService.Create(dtos.SpaceCreateInput{Title: "Docker", Username: "docker", Description: "This is a test description"}, user2).Data.(dtos.SpaceOutput)

	// subscribe to spaces
	spaceService.Subscribe(space2.ID, user1)

	// create some posts
	post1 := postService.Create(dtos.PostInput{Title: "My First Post", Content: "This a the first post in the community"}, space1.ID, user1).Data.(dtos.PostOutput)
	post2 := postService.Create(dtos.PostInput{Title: "What is Docker", Content: "You knew it."}, space2.ID, user1).Data.(dtos.PostOutput)
	postService.Create(dtos.PostInput{Title: "Docker in Production", Content: "I'm not ganna expalin that :)"}, space2.ID, user2)

	// create some comments
	comment1 := commentService.Create(dtos.CommentInput{Content: "So this is the first comment"}, post1.ID, user2).Data.(dtos.CommentOutput)
	commentService.Create(dtos.CommentInput{Content: "You are right!", ParentID: comment1.ID}, post1.ID, user1)
	comment3 := commentService.Create(dtos.CommentInput{Content: "Are you kidding me?"}, post2.ID, user2).Data.(dtos.CommentOutput)

	// vote
	postService.Vote(post1.ID, true, user2)
	commentService.Vote(comment1.ID, true, user2)
	commentService.Vote(comment3.ID, false, user1)

}
