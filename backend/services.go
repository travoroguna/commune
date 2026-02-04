package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// setupServiceRoutes sets up all service-related routes
func setupServiceRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.HandleFunc("/api/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getServicesHandler(db)(w, r)
		case http.MethodPost:
			createServiceRequestHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handle /api/services/{id}
	mux.HandleFunc("/api/services/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getServiceByIDHandler(db)(w, r)
		case http.MethodPut:
			updateServiceRequestHandler(db)(w, r)
		case http.MethodDelete:
			deleteServiceRequestHandler(db)(w, r)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

// getServicesHandler handles GET /api/services with filters
func getServicesHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := db.Model(&ServiceRequest{})

		// Parse query parameters for filtering
		category := r.URL.Query().Get("category")
		status := r.URL.Query().Get("status")
		communityID := r.URL.Query().Get("community_id")
		search := r.URL.Query().Get("search")

		// Apply filters
		if category != "" {
			query = query.Where("category = ?", category)
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if communityID != "" {
			query = query.Where("community_id = ?", communityID)
		}
		if search != "" {
			searchPattern := "%" + search + "%"
			query = query.Where("title LIKE ? OR description LIKE ?", searchPattern, searchPattern)
		}

		// Preload related data
		var services []ServiceRequest
		if err := query.
			Preload("Requester").
			Preload("Community").
			Preload("ServiceOffers").
			Order("created_at DESC").
			Find(&services).Error; err != nil {
			writeError(w, "Failed to fetch services", http.StatusInternalServerError)
			return
		}

		writeJSON(w, services, http.StatusOK)
	}
}

// getServiceByIDHandler handles GET /api/services/{id}
func getServiceByIDHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		idStr := strings.TrimPrefix(r.URL.Path, "/api/services/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		var service ServiceRequest
		if err := db.
			Preload("Requester").
			Preload("Community").
			Preload("ServiceOffers").
			Preload("ServiceOffers.Provider").
			Preload("Comments").
			Preload("Comments.Author").
			First(&service, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Service not found", http.StatusNotFound)
				return
			}
			writeError(w, "Failed to fetch service", http.StatusInternalServerError)
			return
		}

		writeJSON(w, service, http.StatusOK)
	}
}

// CreateServiceRequestInput represents the input for creating a service request
type CreateServiceRequestInput struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	CommunityID uint    `json:"community_id"`
	Budget      float64 `json:"budget"`
}

// createServiceRequestHandler handles POST /api/services
func createServiceRequestHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from session
		userID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var input CreateServiceRequestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if input.Title == "" || input.Description == "" || input.CommunityID == 0 {
			writeError(w, "Title, description, and community_id are required", http.StatusBadRequest)
			return
		}

		// Create the service request
		service := ServiceRequest{
			Title:       input.Title,
			Description: input.Description,
			Category:    input.Category,
			RequesterID: userID,
			CommunityID: input.CommunityID,
			Status:      "open",
			Budget:      input.Budget,
		}

		if err := db.Create(&service).Error; err != nil {
			writeError(w, "Failed to create service request", http.StatusInternalServerError)
			return
		}

		// Load relationships
		db.Preload("Requester").Preload("Community").First(&service, service.ID)

		writeJSON(w, service, http.StatusCreated)
	}
}

// UpdateServiceRequestInput represents the input for updating a service request
type UpdateServiceRequestInput struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Category    *string  `json:"category"`
	Status      *string  `json:"status"`
	Budget      *float64 `json:"budget"`
}

// updateServiceRequestHandler handles PUT /api/services/{id}
func updateServiceRequestHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from session
		userID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch current user to check role
		var currentUser User
		if err := db.First(&currentUser, userID).Error; err != nil {
			writeError(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Extract ID from path
		idStr := strings.TrimPrefix(r.URL.Path, "/api/services/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		// Fetch the service request
		var service ServiceRequest
		if err := db.First(&service, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Service not found", http.StatusNotFound)
				return
			}
			writeError(w, "Failed to fetch service", http.StatusInternalServerError)
			return
		}

		// Check authorization (only requester or admin can update)
		if service.RequesterID != userID && currentUser.Role != RoleSuperAdmin && currentUser.Role != RoleAdmin {
			writeError(w, "Forbidden", http.StatusForbidden)
			return
		}

		var input UpdateServiceRequestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update fields
		if input.Title != nil {
			service.Title = *input.Title
		}
		if input.Description != nil {
			service.Description = *input.Description
		}
		if input.Category != nil {
			service.Category = *input.Category
		}
		// Update status with validation
		if input.Status != nil {
			newStatus := *input.Status
			// Validate state transitions
			validTransitions := map[string][]string{
				"open":        {"in_progress", "cancelled"},
				"in_progress": {"completed", "cancelled"},
				"completed":   {}, // Cannot transition from completed
				"cancelled":   {}, // Cannot transition from cancelled
			}
			
			allowedNextStates, exists := validTransitions[service.Status]
			if !exists {
				writeError(w, "Invalid current status", http.StatusBadRequest)
				return
			}
			
			// Check if transition is valid
			validTransition := false
			for _, allowed := range allowedNextStates {
				if newStatus == allowed {
					validTransition = true
					break
				}
			}
			
			if !validTransition && newStatus != service.Status {
				writeError(w, fmt.Sprintf("Invalid status transition from %s to %s", service.Status, newStatus), http.StatusBadRequest)
				return
			}
			
			service.Status = newStatus
		}
		if input.Budget != nil {
			service.Budget = *input.Budget
		}

		if err := db.Save(&service).Error; err != nil {
			writeError(w, "Failed to update service request", http.StatusInternalServerError)
			return
		}

		// Load relationships
		db.Preload("Requester").Preload("Community").First(&service, service.ID)

		writeJSON(w, service, http.StatusOK)
	}
}

// deleteServiceRequestHandler handles DELETE /api/services/{id}
func deleteServiceRequestHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from session
		userID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch current user to check role
		var currentUser User
		if err := db.First(&currentUser, userID).Error; err != nil {
			writeError(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Extract ID from path
		idStr := strings.TrimPrefix(r.URL.Path, "/api/services/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid service ID", http.StatusBadRequest)
			return
		}

		// Fetch the service request
		var service ServiceRequest
		if err := db.First(&service, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Service not found", http.StatusNotFound)
				return
			}
			writeError(w, "Failed to fetch service", http.StatusInternalServerError)
			return
		}

		// Check authorization (only requester or admin can delete)
		if service.RequesterID != userID && currentUser.Role != RoleSuperAdmin && currentUser.Role != RoleAdmin {
			writeError(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Soft delete
		if err := db.Delete(&service).Error; err != nil {
			writeError(w, "Failed to delete service request", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
