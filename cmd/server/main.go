package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kemit/trip-planner/internal/config"
	"github.com/kemit/trip-planner/internal/database"
	"github.com/kemit/trip-planner/internal/handlers"
	"github.com/kemit/trip-planner/internal/middleware"
	"github.com/kemit/trip-planner/internal/repository"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

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

	// ── Repositories ──────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	placeRepo := repository.NewPlaceRepository(db)
	tripRepo := repository.NewTripRepository(db)
	chatRepo := repository.NewChatRepository(db)
	savedPlaceRepo := repository.NewSavedPlaceRepository(db)

	// ── Services ──────────────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, jwtManager)
	userService := services.NewUserService(userRepo, savedPlaceRepo, chatRepo)
	placeService := services.NewPlaceService(placeRepo)
	tripService := services.NewTripService(tripRepo, placeRepo, aiClient)
	chatService := services.NewChatService(chatRepo, aiClient)
	savedPlaceService := services.NewSavedPlaceService(savedPlaceRepo, placeRepo)
	cloudinaryService := services.NewCloudinaryService(cfg.CloudinaryURL)

	// ── Handlers ──────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, authService, savedPlaceService, cloudinaryService)
	tripHandler := handlers.NewTripHandler(tripService)
	placeHandler := handlers.NewPlaceHandler(placeService, savedPlaceService, cloudinaryService)
	chatHandler := handlers.NewChatHandler(chatService)

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
			auth.POST("/refresh", authHandler.RefreshToken) // public: must work with expired access tokens
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
			users.PUT("/profile", userHandler.UpdateProfile)
			users.PUT("/preferences", userHandler.UpdatePreferences)
			users.PUT("/password", userHandler.ChangePassword)
			users.PUT("/avatar", userHandler.UpdateAvatar)
			users.GET("/stats", userHandler.GetStats)
			users.GET("/saved-places", userHandler.GetSavedPlaces)
		}

		// ── Place Routes (Public) ────────────────────────────
		places := api.Group("/places")
		{
			places.GET("", placeHandler.ListPlaces)
			places.GET("/popular", placeHandler.PopularPlaces)
			places.GET("/search", placeHandler.SearchPlaces)
			places.GET("/categories", placeHandler.ListCategories)
			places.GET("/trending", placeHandler.TrendingSearches)
			places.GET("/:id", placeHandler.GetPlace)
			places.POST("/:id/photo", placeHandler.UploadPlacePhoto) // Optionally protected
		}

		// ── Place Routes (Protected — Save/Unsave) ───────────
		placesProtected := api.Group("/places")
		placesProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			placesProtected.POST("/:id/save", placeHandler.SavePlace)
			placesProtected.DELETE("/:id/save", placeHandler.UnsavePlace)
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
			trips.PUT("/:id/activities/:activityId", tripHandler.UpdateActivity)
			trips.DELETE("/:id/activities/:activityId", tripHandler.DeleteActivity)
		}

		// ── Chat Routes (Protected — Ask WAYNX) ─────────────
		chat := api.Group("/chat")
		chat.Use(middleware.AuthMiddleware(jwtManager))
		{
			chat.POST("", chatHandler.SendMessage)
			chat.GET("/history", chatHandler.ListSessions)
			chat.GET("/:id", chatHandler.GetSession)
			chat.DELETE("/:id", chatHandler.DeleteSession)
		}
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("🚀 Trip Planner API starting on http://localhost%s", addr)
	log.Printf("📋 Health check: http://localhost%s/health", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
