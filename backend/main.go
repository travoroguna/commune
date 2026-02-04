package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

	// Setup HTTP server
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/health", healthHandler)

	// Auth routes
	mux.HandleFunc("/api/auth/login", loginHandler(db))
	mux.HandleFunc("/api/auth/logout", logoutHandler())
	mux.HandleFunc("/api/auth/me", getCurrentUserHandler(db))
	mux.HandleFunc("/api/auth/first-boot", checkFirstBootHandler(db))
	mux.HandleFunc("/api/auth/setup-super-user", setupSuperUserHandler(db))

	// User routes - need router to handle different methods and paths
	setupUserRoutes(mux, db)

	// Community routes
	setupCommunityRoutes(mux, db)

	// Join request routes
	setupJoinRequestRoutes(mux, db)

	// Service request and offer routes
	setupServiceRequestRoutes(mux, db)

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
	mux.Handle("/", viteHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// Route setup functions
func setupUserRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getUsersHandler(db)(w, r)
		case http.MethodPost:
			createUserHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/users/change-password", changePasswordHandler(db))

	// Handle /api/users/{id} and /api/users/{id}/communities
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if it's /api/users/{id}/communities
		if strings.Contains(path, "/communities") {
			getUserCommunitiesHandler(db)(w, r)
			return
		}

		// Otherwise it's /api/users/{id}
		switch r.Method {
		case http.MethodGet:
			getUserByIDHandler(db)(w, r)
		case http.MethodPut:
			updateUserHandler(db)(w, r)
		case http.MethodDelete:
			deleteUserHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func setupCommunityRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.HandleFunc("/api/communities", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getCommunitiesHandler(db)(w, r)
		case http.MethodPost:
			createCommunityHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handle /api/communities/{id}, /api/communities/{id}/members, /api/communities/{id}/join-requests
	mux.HandleFunc("/api/communities/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check if it's join-requests endpoint
		if strings.Contains(path, "/join-requests") {
			getCommunityJoinRequestsHandler(db)(w, r)
			return
		}

		// Check if it's members endpoint
		if strings.Contains(path, "/members") {
			parts := strings.Split(strings.TrimPrefix(path, "/api/communities/"), "/")
			if len(parts) >= 3 {
				// /api/communities/{id}/members/{userId}
				switch r.Method {
				case http.MethodDelete:
					removeCommunityMemberHandler(db)(w, r)
				case http.MethodPut:
					updateCommunityMemberRoleHandler(db)(w, r)
				default:
					writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else {
				// /api/communities/{id}/members
				switch r.Method {
				case http.MethodGet:
					getCommunityMembersHandler(db)(w, r)
				case http.MethodPost:
					addCommunityMemberHandler(db)(w, r)
				default:
					writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}
			return
		}

		// Otherwise it's /api/communities/{id}
		switch r.Method {
		case http.MethodGet:
			getCommunityByIDHandler(db)(w, r)
		case http.MethodPut:
			updateCommunityHandler(db)(w, r)
		case http.MethodDelete:
			deleteCommunityHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func setupJoinRequestRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.HandleFunc("/api/join-requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getJoinRequestsHandler(db)(w, r)
		case http.MethodPost:
			createJoinRequestHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handle /api/join-requests/{id}/approve and /api/join-requests/{id}/reject
	mux.HandleFunc("/api/join-requests/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/approve") {
			approveJoinRequestHandler(db)(w, r)
		} else if strings.HasSuffix(path, "/reject") {
			rejectJoinRequestHandler(db)(w, r)
		} else {
			writeError(w, "Not found", http.StatusNotFound)
		}
	})
}
