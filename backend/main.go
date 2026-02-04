package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	mux.HandleFunc("/api/users", usersHandler(db))

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
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
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

func usersHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			var users []User
			if err := db.Find(&users).Error; err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Failed to fetch users"}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"users":[]}`))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
		}
	}
}
