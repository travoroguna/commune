package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// serviceRequestsHandler handles GET (list) and POST (create) for service requests
func serviceRequestsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		switch c.Request.Method {
		case http.MethodGet:
			listServiceRequests(c, db)
		case http.MethodPost:
			createServiceRequest(c, db)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		}
	}
}

// listServiceRequests handles GET /api/service-requests
func listServiceRequests(c *gin.Context, db *gorm.DB) {
	// Get query parameters
	communityIDStr := c.Query("community_id")
	status := c.Query("status")
	category := c.Query("category")

	// Build query
	query := db.Model(&ServiceRequest{}).
		Preload("Requester").
		Preload("Community").
		Preload("ServiceOffers").
		Preload("ServiceOffers.Provider")

	// Filter by community if specified
	if communityIDStr != "" {
		communityID, err := strconv.ParseUint(communityIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community_id"})
			return
		}
		query = query.Where("community_id = ?", communityID)
	}

	// Filter by status if specified
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by category if specified
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var requests []ServiceRequest
	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service requests"})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// createServiceRequest handles POST /api/service-requests
func createServiceRequest(c *gin.Context, db *gorm.DB) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var input struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		CommunityID uint    `json:"community_id"`
		Budget      float64 `json:"budget"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Validate required fields
	if input.Title == "" || input.Description == "" || input.CommunityID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Title, description, and community_id are required"})
		return
	}

	// Create service request
	request := ServiceRequest{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		RequesterID: userID,
		CommunityID: input.CommunityID,
		Status:      "open",
		Budget:      input.Budget,
	}

	if err := db.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create service request"})
		return
	}

	// Reload with associations
	if err := db.Preload("Requester").Preload("Community").First(&request, request.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load created request"})
		return
	}

	c.JSON(http.StatusCreated, request)
}

// serviceRequestDetailHandler handles GET, PUT, DELETE for a specific service request
func serviceRequestDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		requestID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service request ID"})
			return
		}

		// Handle accept offer endpoint - check if this is the accept-offer sub-route
		if c.Param("action") == "accept-offer" {
			acceptServiceOffer(c, db, uint(requestID))
			return
		}

		switch c.Request.Method {
		case http.MethodGet:
			getServiceRequest(c, db, uint(requestID))
		case http.MethodPut:
			updateServiceRequest(c, db, uint(requestID))
		case http.MethodDelete:
			deleteServiceRequest(c, db, uint(requestID))
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		}
	}
}

// getServiceRequest handles GET /api/service-requests/:id
func getServiceRequest(c *gin.Context, db *gorm.DB, requestID uint) {
	var request ServiceRequest
	if err := db.Preload("Requester").
		Preload("Community").
		Preload("ServiceOffers").
		Preload("ServiceOffers.Provider").
		Preload("AcceptedOffer").
		Preload("AcceptedOffer.Provider").
		First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service request"})
		}
		return
	}

	c.JSON(http.StatusOK, request)
}

// updateServiceRequest handles PUT /api/service-requests/:id
func updateServiceRequest(c *gin.Context, db *gorm.DB, requestID uint) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userRole, _ := c.Get("userRole")

	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service request"})
		}
		return
	}

	// Only requester can update
	if request.RequesterID != userID && userRole != RoleSuperAdmin && userRole != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	var input struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Category    *string  `json:"category"`
		Budget      *float64 `json:"budget"`
		Status      *string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Category != nil {
		updates["category"] = *input.Category
	}
	if input.Budget != nil {
		updates["budget"] = *input.Budget
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}

	if err := db.Model(&request).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update service request"})
		return
	}

	// Reload with associations
	if err := db.Preload("Requester").Preload("Community").First(&request, requestID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load updated request"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// deleteServiceRequest handles DELETE /api/service-requests/:id
func deleteServiceRequest(c *gin.Context, db *gorm.DB, requestID uint) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	userRole, _ := c.Get("userRole")

	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service request"})
		}
		return
	}

	// Only requester or admin can delete
	if request.RequesterID != userID && userRole != RoleSuperAdmin && userRole != RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	// Soft delete
	if err := db.Delete(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete service request"})
		return
	}

	c.Status(http.StatusNoContent)
}

// acceptServiceOffer handles PUT /api/service-requests/:id/accept-offer
func acceptServiceOffer(c *gin.Context, db *gorm.DB, requestID uint) {
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		return
	}

	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var input struct {
		OfferID uint `json:"offer_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if input.OfferID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "offer_id is required"})
		return
	}

	// Verify request exists and user is requester
	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service request"})
		}
		return
	}

	if request.RequesterID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized: only requester can accept offers"})
		return
	}

	// Verify offer exists and belongs to this request
	var offer ServiceOffer
	if err := db.First(&offer, input.OfferID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch offer"})
		}
		return
	}

	if offer.ServiceRequestID != requestID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Offer does not belong to this request"})
		return
	}

	// Update request and offer in transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// Update service request
		if err := tx.Model(&request).Updates(map[string]interface{}{
			"accepted_offer_id": input.OfferID,
			"status":            "in_progress",
		}).Error; err != nil {
			return err
		}

		// Update accepted offer status
		if err := tx.Model(&offer).Update("status", "accepted").Error; err != nil {
			return err
		}

		// Reject other offers
		if err := tx.Model(&ServiceOffer{}).
			Where("service_request_id = ? AND id != ?", requestID, input.OfferID).
			Update("status", "rejected").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to accept offer"})
		return
	}

	// Reload request with associations
	if err := db.Preload("Requester").
		Preload("Community").
		Preload("AcceptedOffer").
		Preload("AcceptedOffer.Provider").
		First(&request, requestID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load updated request"})
		return
	}

	c.JSON(http.StatusOK, request)
}

// serviceOffersHandler handles GET (list) and POST (create) for service offers
func serviceOffersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		switch c.Request.Method {
		case http.MethodGet:
			listServiceOffers(c, db)
		case http.MethodPost:
			createServiceOffer(c, db)
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		}
	}
}

// listServiceOffers handles GET /api/service-offers
func listServiceOffers(c *gin.Context, db *gorm.DB) {
	userID, _ := getCurrentUser(c)

	// Get query parameters
	serviceRequestIDStr := c.Query("service_request_id")
	myOffers := c.Query("my_offers") == "true"
	providerIDStr := c.Query("provider_id")

	query := db.Model(&ServiceOffer{}).
		Preload("Provider").
		Preload("ServiceRequest").
		Preload("ServiceRequest.Requester").
		Preload("ServiceRequest.Community")

	// Filter by service request if specified
	if serviceRequestIDStr != "" {
		serviceRequestID, err := strconv.ParseUint(serviceRequestIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service_request_id"})
			return
		}
		query = query.Where("service_request_id = ?", serviceRequestID)
	}

	// Filter by current user's offers if requested
	if myOffers {
		query = query.Where("provider_id = ?", userID)
	}

	// Filter by provider ID if specified
	if providerIDStr != "" {
		providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid provider_id"})
			return
		}
		query = query.Where("provider_id = ?", providerID)
	}

	var offers []ServiceOffer
	if err := query.Order("created_at DESC").Find(&offers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service offers"})
		return
	}

	c.JSON(http.StatusOK, offers)
}

// createServiceOffer handles POST /api/service-offers
func createServiceOffer(c *gin.Context, db *gorm.DB) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var input struct {
		ServiceRequestID  uint    `json:"service_request_id"`
		Description       string  `json:"description"`
		ProposedPrice     float64 `json:"proposed_price"`
		EstimatedDuration string  `json:"estimated_duration"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Validate required fields
	if input.ServiceRequestID == 0 || input.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "service_request_id and description are required"})
		return
	}

	// Verify service request exists and is open
	var request ServiceRequest
	if err := db.First(&request, input.ServiceRequestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service request not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service request"})
		}
		return
	}

	if request.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot create offer for non-open requests"})
		return
	}

	// Create service offer
	offer := ServiceOffer{
		ServiceRequestID:  input.ServiceRequestID,
		ProviderID:        userID,
		Description:       input.Description,
		ProposedPrice:     input.ProposedPrice,
		EstimatedDuration: input.EstimatedDuration,
		Status:            "pending",
	}

	if err := db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create service offer"})
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offer.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load created offer"})
		return
	}

	c.JSON(http.StatusCreated, offer)
}

// serviceOfferDetailHandler handles GET, PUT, DELETE for a specific service offer
func serviceOfferDetailHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		offerID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid service offer ID"})
			return
		}

		// Handle withdraw endpoint
		if c.Param("action") == "withdraw" {
			withdrawServiceOffer(c, db, uint(offerID))
			return
		}

		switch c.Request.Method {
		case http.MethodGet:
			getServiceOffer(c, db, uint(offerID))
		case http.MethodPut:
			updateServiceOffer(c, db, uint(offerID))
		case http.MethodDelete:
			deleteServiceOffer(c, db, uint(offerID))
		default:
			c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		}
	}
}

// withdrawServiceOffer handles POST /api/service-offers/:id/withdraw
func withdrawServiceOffer(c *gin.Context, db *gorm.DB, offerID uint) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"message": "Method not allowed"})
		return
	}

	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service offer"})
		}
		return
	}

	// Only provider can withdraw
	if offer.ProviderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	// Cannot withdraw accepted offers
	if offer.Status == "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot withdraw accepted offers"})
		return
	}

	// Update status to withdrawn
	if err := db.Model(&offer).Update("status", "withdrawn").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to withdraw service offer"})
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load updated offer"})
		return
	}

	c.JSON(http.StatusOK, offer)
}

// getServiceOffer handles GET /api/service-offers/:id
func getServiceOffer(c *gin.Context, db *gorm.DB, offerID uint) {
	var offer ServiceOffer
	if err := db.Preload("Provider").
		Preload("ServiceRequest").
		Preload("ServiceRequest.Requester").
		First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service offer"})
		}
		return
	}

	c.JSON(http.StatusOK, offer)
}

// updateServiceOffer handles PUT /api/service-offers/:id
func updateServiceOffer(c *gin.Context, db *gorm.DB, offerID uint) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service offer"})
		}
		return
	}

	// Only provider can update
	if offer.ProviderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	var input struct {
		Description       *string  `json:"description"`
		ProposedPrice     *float64 `json:"proposed_price"`
		EstimatedDuration *string  `json:"estimated_duration"`
		Status            *string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.ProposedPrice != nil {
		updates["proposed_price"] = *input.ProposedPrice
	}
	if input.EstimatedDuration != nil {
		updates["estimated_duration"] = *input.EstimatedDuration
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}

	if err := db.Model(&offer).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update service offer"})
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to load updated offer"})
		return
	}

	c.JSON(http.StatusOK, offer)
}

// deleteServiceOffer handles DELETE /api/service-offers/:id (withdraw offer)
func deleteServiceOffer(c *gin.Context, db *gorm.DB, offerID uint) {
	userID, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "Service offer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch service offer"})
		}
		return
	}

	// Only provider can delete
	if offer.ProviderID != userID {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	// Cannot delete accepted offers
	if offer.Status == "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot withdraw accepted offers"})
		return
	}

	// Soft delete
	if err := db.Delete(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete service offer"})
		return
	}

	c.Status(http.StatusNoContent)
}
