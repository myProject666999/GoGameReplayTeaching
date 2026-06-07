package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/goreplay/backend/config"
	"github.com/goreplay/backend/database"
	"github.com/goreplay/backend/handlers"
	"github.com/goreplay/backend/middleware"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.CORS())

	userHandler := handlers.NewUserHandler(cfg)
	gameHandler := handlers.NewGameHandler()
	commentHandler := handlers.NewCommentHandler()
	markerHandler := handlers.NewMarkerHandler()
	problemHandler := handlers.NewProblemHandler()

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", userHandler.Register)
	auth.POST("/login", userHandler.Login)

	users := api.Group("/users")
	users.GET("", userHandler.List)
	users.GET("/:id", userHandler.Get)
	users.GET("/me", userHandler.Me, middleware.JWTAuth(cfg))

	games := api.Group("/games")
	games.GET("", gameHandler.List)
	games.GET("/:id", gameHandler.Get)
	games.POST("/parse", gameHandler.ParseSGF)
	games.POST("", gameHandler.Create, middleware.JWTAuth(cfg))
	games.PUT("/:id", gameHandler.Update, middleware.JWTAuth(cfg))
	games.DELETE("/:id", gameHandler.Delete, middleware.JWTAuth(cfg))

	comments := api.Group("/comments")
	comments.GET("/game/:game_id", commentHandler.ListByGame)
	comments.POST("", commentHandler.Create, middleware.JWTAuth(cfg))
	comments.PUT("/:id", commentHandler.Update, middleware.JWTAuth(cfg))
	comments.DELETE("/:id", commentHandler.Delete, middleware.JWTAuth(cfg))

	markers := api.Group("/markers")
	markers.GET("/game/:game_id", markerHandler.ListByGame)
	markers.POST("", markerHandler.Create, middleware.JWTAuth(cfg))
	markers.DELETE("/:id", markerHandler.Delete, middleware.JWTAuth(cfg))

	problems := api.Group("/problems")
	problems.GET("", problemHandler.List)
	problems.GET("/:id", problemHandler.Get)
	problems.POST("", problemHandler.Create, middleware.JWTAuth(cfg))
	problems.PUT("/:id", problemHandler.Update, middleware.JWTAuth(cfg))
	problems.DELETE("/:id", problemHandler.Delete, middleware.JWTAuth(cfg))
	problems.POST("/:id/attempt", problemHandler.Attempt, middleware.JWTAuth(cfg))
	problems.GET("/:id/attempts", problemHandler.ListAttempts, middleware.JWTAuth(cfg))

	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	log.Printf("Server starting on %s", cfg.ServerPort)
	if err := e.Start(cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
