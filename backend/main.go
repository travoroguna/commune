package main

import (
	"log"
	"net/http"
	"os"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"github.com/olivere/vite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Example User model
type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"uniqueIndex"`
}

func main() {
	// Initialize database
	db, err := gorm.Open(sqlite.Open("commune.db"), &gorm.Config{})
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
	var viteHandler http.Handler
	if mode == "production" {
		// In production, serve the built static files
		log.Println("Running in PRODUCTION mode")
		viteHandler = vite.NewStatic("../frontend/dist")
	} else {
		// In development, proxy to Vite dev server
		log.Println("Running in DEVELOPMENT mode")
		viteHandler = vite.NewProxy("http://localhost:5173")
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
				return tx.AutoMigrate(&User{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("users")
			},
		},
	})

	return m.Migrate()
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
