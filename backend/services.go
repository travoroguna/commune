package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getServicesHandler handles GET /api/services with filters
func getServicesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&ServiceRequest{})

		// Parse query parameters for filtering
		category := c.Query("category")
		status := c.Query("status")
		communityID := c.Query("community_id")
		search := c.Query("search")

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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch services"})
			return
		}

		c.JSON(http.StatusOK, services)
	}
}

// getServiceByIDHandler handles GET /api/services/{id}
func getServiceByIDHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service ID"})
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
				c.JSON(http.StatusNotFound, gin.H{"message": "Service not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service"})
			return
		}

		c.JSON(http.StatusOK, service)
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
func createServiceRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		var input CreateServiceRequestInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		// Validate required fields
		if input.Title == "" || input.Description == "" || input.CommunityID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Title, description, and community_id are required"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create service request"})
			return
		}

		// Load relationships
		db.Preload("Requester").Preload("Community").First(&service, service.ID)

		c.JSON(http.StatusCreated, service)
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
func updateServiceRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		// Fetch current user to check role
		var currentUser User
		if err := db.First(&currentUser, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service ID"})
			return
		}

		// Fetch the service request
		var service ServiceRequest
		if err := db.First(&service, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Service not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service"})
			return
		}

		// Check authorization (only requester or admin can update)
		if service.RequesterID != userID && currentUser.Role != RoleSuperAdmin && currentUser.Role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			return
		}

		var input UpdateServiceRequestInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
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
				c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid current status"})
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
				c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("Invalid status transition from %s to %s", service.Status, newStatus)})
				return
			}

			service.Status = newStatus
		}
		if input.Budget != nil {
			service.Budget = *input.Budget
		}

		if err := db.Save(&service).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update service request"})
			return
		}

		// Load relationships
		db.Preload("Requester").Preload("Community").First(&service, service.ID)

		c.JSON(http.StatusOK, service)
	}
}

// deleteServiceRequestHandler handles DELETE /api/services/{id}
func deleteServiceRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		// Fetch current user to check role
		var currentUser User
		if err := db.First(&currentUser, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service ID"})
			return
		}

		// Fetch the service request
		var service ServiceRequest
		if err := db.First(&service, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Service not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service"})
			return
		}

		// Check authorization (only requester or admin can delete)
		if service.RequesterID != userID && currentUser.Role != RoleSuperAdmin && currentUser.Role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			return
		}

		// Soft delete
		if err := db.Delete(&service).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete service request"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
