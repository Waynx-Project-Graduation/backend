package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	_ "github.com/kemit/trip-planner/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/kemit/trip-planner/internal/config"
	"github.com/kemit/trip-planner/internal/database"
	"github.com/kemit/trip-planner/internal/handlers"
	"github.com/kemit/trip-planner/internal/middleware"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

// @title           Kemit Trip Planner API
// @version         1.0
// @description     This is the backend API for the Kemit Trip Planner application.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Connect to database
	db := database.Connect(cfg)
	database.AutoMigrate(db)

	// Initialize JWT manager
	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTRefreshExpiry)

	// Initialize AI client
	aiClient := services.NewAIClient(cfg.GeminiAPIKey, cfg.WAYNXBaseURL, cfg.AITimeoutSeconds)
	defer aiClient.Close()

	// ── Repositories ──────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	tripRepo := repository.NewTripRepository(db)
	chatRepo := repository.NewChatRepository(db)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	expenseRepo := repository.NewExpenseRepository(db)

	// ── Services ──────────────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, jwtManager)
	userService := services.NewUserService(userRepo, savedPlaceRepo, chatRepo)
	placeService := services.NewPlaceService(placeRepo, aiClient)
	tripService := services.NewTripService(tripRepo, placeRepo, aiClient)
	chatService := services.NewChatService(chatRepo, userRepo, savedPlaceRepo, placeRepo, aiClient)
	savedPlaceService := services.NewSavedPlaceService(savedPlaceRepo, placeRepo)
	cloudinaryService := services.NewCloudinaryService(cfg.CloudinaryURL)
	notifService := services.NewNotificationService(notifRepo)
	reviewService := services.NewReviewService(reviewRepo, placeRepo)
	expenseService := services.NewExpenseService(expenseRepo, tripRepo)

	// Wire the gamification points awarder into the action-producing services.
	tripService.SetPointsAwarder(userService)
	reviewService.SetPointsAwarder(userService)
	savedPlaceService.SetPointsAwarder(userService)

	// ── Handlers ──────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, authService, savedPlaceService, cloudinaryService)
	tripHandler := handlers.NewTripHandler(tripService)
	placeHandler := handlers.NewPlaceHandler(placeService, savedPlaceService)
	chatHandler := handlers.NewChatHandler(chatService)
	notifHandler := handlers.NewNotificationHandler(notifService)
	reviewHandler := handlers.NewReviewHandler(reviewService)
	expenseHandler := handlers.NewExpenseHandler(expenseService)

	// ── Router Setup ──────────────────────────────────────────────────
	r := gin.Default()

	// Global middleware
	r.Use(middleware.ErrorHandler())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false, // wildcard origin does not allow credentials
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "trip-planner-api"})
	})

	api := r.Group("/api")
	{
		// ── Auth Routes (Public) ──────────────────────────────
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/google", authHandler.GoogleAuth)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// ── Auth Routes (Protected) ──────────────────────────
		authProtected := api.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			authProtected.GET("/me", authHandler.Me)
			authProtected.POST("/logout", authHandler.Logout)
		}

		// ── User Routes (Protected) ──────────────────────────
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.PUT("/preferences", userHandler.UpdatePreferences)
			users.PUT("/password", userHandler.ChangePassword)
			users.PUT("/avatar", userHandler.UpdateAvatar)
			users.GET("/stats", userHandler.GetStats)
			users.GET("/saved-places", userHandler.GetSavedPlaces)
			users.DELETE("/account", userHandler.DeleteAccount)
		}

		// ── Place Routes (Public) ────────────────────────────
		places := api.Group("/places")
		{
			places.POST("/recommend", placeHandler.RecommendPlaces)
			places.GET("", placeHandler.ListPlaces)
			places.GET("/popular", placeHandler.PopularPlaces)
			places.GET("/search", placeHandler.SearchPlaces)
			places.GET("/categories", placeHandler.ListCategories)
			places.GET("/cities", placeHandler.ListCities)
			places.GET("/trending", placeHandler.TrendingSearches)
			places.GET("/:id", placeHandler.GetPlace)
			places.GET("/:id/reviews", reviewHandler.ListReviews)
		}

		// ── Place Routes (Protected) ─────────────────────────
		placesProtected := api.Group("/places")
		placesProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			placesProtected.POST("/:id/save", placeHandler.SavePlace)
			placesProtected.DELETE("/:id/save", placeHandler.UnsavePlace)
			placesProtected.GET("/:id/save", placeHandler.IsSaved)
			placesProtected.POST("/:id/reviews", reviewHandler.CreateReview)
			placesProtected.PUT("/:id/reviews/:reviewId", reviewHandler.UpdateReview)
			placesProtected.DELETE("/:id/reviews/:reviewId", reviewHandler.DeleteReview)
		}

		// ── Trip Routes (Protected) ──────────────────────────
		trips := api.Group("/trips")
		trips.Use(middleware.AuthMiddleware(jwtManager))
		{
			trips.POST("", tripHandler.CreateTrip)
			trips.GET("", tripHandler.ListTrips)
			trips.GET("/:id", tripHandler.GetTrip)
			trips.PUT("/:id", tripHandler.UpdateTrip)
			trips.DELETE("/:id", tripHandler.DeleteTrip)
			trips.POST("/:id/regenerate", tripHandler.RegenerateItinerary)

			// Trip expenses
			trips.POST("/:id/expenses", expenseHandler.CreateExpense)
			trips.GET("/:id/expenses", expenseHandler.ListExpenses)
			trips.PUT("/:id/expenses/:expenseId", expenseHandler.UpdateExpense)
			trips.DELETE("/:id/expenses/:expenseId", expenseHandler.DeleteExpense)
		}

		// ── Chat Routes (Protected) ──────────────────────────
		chat := api.Group("/chat")
		chat.Use(middleware.AuthMiddleware(jwtManager))
		{
			// AI-generating endpoints are rate-limited per user to bound cost:
			// 20 messages/min with a small burst allowance.
			chatLimit := middleware.RateLimitPerUser(20, 5)
			chat.POST("", chatLimit, chatHandler.SendMessage)
			chat.POST("/stream", chatLimit, chatHandler.StreamMessage)
			chat.GET("/history", chatHandler.ListSessions)
			chat.GET("/:id", chatHandler.GetSession)
			chat.PUT("/:id", chatHandler.RenameSession)
			chat.DELETE("/:id", chatHandler.DeleteSession)
		}

		// ── Notification Routes (Protected) ──────────────────
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(jwtManager))
		{
			notifications.GET("", notifHandler.ListNotifications)
			notifications.GET("/unread-count", notifHandler.UnreadCount)
			notifications.PUT("/:id/read", notifHandler.MarkAsRead)
			notifications.PUT("/read-all", notifHandler.MarkAllAsRead)
		}
	}

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	addr := ":" + cfg.Port
	log.Printf("🚀 Trip Planner API starting on http://localhost:%s\n", cfg.Port)
	log.Printf("📋 Health check: http://localhost:%s/health\n", cfg.Port)
	log.Printf("📚 Swagger Documentation: http://localhost:%s/swagger/index.html\n", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
