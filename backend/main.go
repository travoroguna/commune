package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/olivere/vite"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Note: Models are now defined in models.go

func main() {
	// Initialize database
	db, err := initDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Get mode from environment (default to development)
	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "development"
	}

	// Set Gin mode
	if mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.Default()

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", loginHandler(db))
			auth.POST("/logout", logoutHandler())
			auth.GET("/me", getCurrentUserHandler(db))
			auth.GET("/first-boot", checkFirstBootHandler(db))
			auth.POST("/setup-super-user", setupSuperUserHandler(db))
		}

		// User routes
		setupUserRoutes(api, db)

		// Community routes
		setupCommunityRoutes(api, db)

		// Join request routes
		setupJoinRequestRoutes(api, db)

		// Service routes (simple services page)
		setupServiceRoutes(api, db)

		// Service request and offer routes (marketplace)
		setupServiceRequestRoutes(api, db)
	}

	// Vite integration for serving frontend
	var viteHandler *vite.Handler
	if mode == "production" {
		// In production, serve the built static files
		log.Println("Running in PRODUCTION mode")
		distFS := os.DirFS("../frontend/dist")
		viteHandler, err = vite.NewHandler(vite.Config{
			FS:    distFS,
			IsDev: false,
		})
		if err != nil {
			log.Fatal("Failed to create vite handler:", err)
		}
	} else {
		// In development, proxy to Vite dev server
		log.Println("Running in DEVELOPMENT mode")
		viteHandler, err = vite.NewHandler(vite.Config{
			FS:      os.DirFS("../frontend"),
			IsDev:   true,
			ViteURL: "http://localhost:5173",
		})
		if err != nil {
			log.Fatal("Failed to create vite handler:", err)
		}
	}

	// Use vite handler for all non-API routes
	router.NoRoute(gin.WrapH(viteHandler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func runMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202402041300",
			Migrate: func(tx *gorm.DB) error {
				// Initial User table (legacy, kept for backwards compatibility)
				type OldUser struct {
					gorm.Model
					Name  string
					Email string `gorm:"uniqueIndex"`
				}
				return tx.AutoMigrate(&OldUser{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("users")
			},
		},
		{
			ID: "202402041301",
			Migrate: func(tx *gorm.DB) error {
				// Create all new tables for the community marketplace
				return tx.AutoMigrate(
					&User{},
					&Community{},
					&UserCommunity{},
					&Post{},
					&ServiceRequest{},
					&ServiceOffer{},
					&Comment{},
					&Rating{},
					&JoinRequest{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					"join_requests",
					"ratings",
					"comments",
					"service_offers",
					"service_requests",
					"posts",
					"user_communities",
					"communities",
					"users",
				)
			},
		},
	})

	if err := m.Migrate(); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

func initDatabase() (*gorm.DB, error) {
	// Check if PostgreSQL connection details are provided
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// If PostgreSQL env vars are set, use PostgreSQL
	if dbHost != "" && dbName != "" && dbUser != "" && dbPassword != "" {
		if dbPort == "" {
			dbPort = "5432"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			dbHost, dbUser, dbPassword, dbName, dbPort)
		log.Println("Connecting to PostgreSQL database...")
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	// Otherwise, fallback to SQLite
	log.Println("Connecting to SQLite database...")
	return gorm.Open(sqlite.Open("commune.db"), &gorm.Config{})
}

// Route setup functions

func setupUserRoutes(api *gin.RouterGroup, db *gorm.DB) {
	users := api.Group("/users")
	{
		users.GET("", getUsersHandler(db))
		users.POST("", createUserHandler(db))
		users.POST("/change-password", changePasswordHandler(db))
		users.GET("/:id", getUserByIDHandler(db))
		users.PUT("/:id", updateUserHandler(db))
		users.DELETE("/:id", deleteUserHandler(db))
		users.GET("/:id/communities", getUserCommunitiesHandler(db))
	}
}

func setupCommunityRoutes(api *gin.RouterGroup, db *gorm.DB) {
	communities := api.Group("/communities")
	{
		communities.GET("", getCommunitiesHandler(db))
		communities.POST("", createCommunityHandler(db))
		communities.GET("/:id", getCommunityByIDHandler(db))
		communities.PUT("/:id", updateCommunityHandler(db))
		communities.DELETE("/:id", deleteCommunityHandler(db))

		// Members
		communities.GET("/:id/members", getCommunityMembersHandler(db))
		communities.POST("/:id/members", addCommunityMemberHandler(db))
		communities.DELETE("/:id/members/:userId", removeCommunityMemberHandler(db))
		communities.PUT("/:id/members/:userId", updateCommunityMemberRoleHandler(db))

		// Join requests
		communities.GET("/:id/join-requests", getCommunityJoinRequestsHandler(db))
	}
}

func setupJoinRequestRoutes(api *gin.RouterGroup, db *gorm.DB) {
	joinRequests := api.Group("/join-requests")
	{
		joinRequests.GET("", getJoinRequestsHandler(db))
		joinRequests.POST("", createJoinRequestHandler(db))
		joinRequests.POST("/:id/approve", approveJoinRequestHandler(db))
		joinRequests.POST("/:id/reject", rejectJoinRequestHandler(db))
	}
}

func setupServiceRoutes(api *gin.RouterGroup, db *gorm.DB) {
	services := api.Group("/services")
	{
		services.GET("", getServicesHandler(db))
		services.POST("", createServiceRequestHandler(db))
		services.GET("/:id", getServiceByIDHandler(db))
		services.PUT("/:id", updateServiceRequestHandler(db))
		services.DELETE("/:id", deleteServiceRequestHandler(db))
	}
}

func setupServiceRequestRoutes(api *gin.RouterGroup, db *gorm.DB) {
	// Service requests - using the compound handlers
	api.GET("/service-requests", serviceRequestsHandler(db))
	api.POST("/service-requests", serviceRequestsHandler(db))
	api.GET("/service-requests/:id", serviceRequestDetailHandler(db))
	api.PUT("/service-requests/:id", serviceRequestDetailHandler(db))
	api.DELETE("/service-requests/:id", serviceRequestDetailHandler(db))

	// Service offers - using the compound handlers
	api.GET("/service-offers", serviceOffersHandler(db))
	api.POST("/service-offers", serviceOffersHandler(db))
	api.GET("/service-offers/:id", serviceOfferDetailHandler(db))
	api.PUT("/service-offers/:id", serviceOfferDetailHandler(db))
	api.DELETE("/service-offers/:id", serviceOfferDetailHandler(db))
}
