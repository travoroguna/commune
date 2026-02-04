package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// setupServiceRequestRoutes sets up routes for service requests and offers
func setupServiceRequestRoutes(mux *http.ServeMux, db *gorm.DB) {
	// Service request routes
	mux.HandleFunc("/api/service-requests", authMiddleware(db)(serviceRequestsHandler(db)))
	mux.HandleFunc("/api/service-requests/", authMiddleware(db)(serviceRequestDetailHandler(db)))
	
	// Service offer routes
	mux.HandleFunc("/api/service-offers", authMiddleware(db)(serviceOffersHandler(db)))
	mux.HandleFunc("/api/service-offers/", authMiddleware(db)(serviceOfferDetailHandler(db)))
}

// getUserFromContext retrieves the user from the request
func getUserFromContext(r *http.Request, db *gorm.DB) (*User, error) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, err
	}
	
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	
	return &user, nil
}

// serviceRequestsHandler handles GET (list) and POST (create) for service requests
func serviceRequestsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromContext(r, db)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			listServiceRequests(w, r, db, user)
		case http.MethodPost:
			createServiceRequest(w, r, db, user)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// listServiceRequests handles GET /api/service-requests
func listServiceRequests(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User) {
	// Get query parameters
	communityIDStr := r.URL.Query().Get("community_id")
	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")

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
			writeError(w, "Invalid community_id", http.StatusBadRequest)
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
		writeError(w, "Failed to fetch service requests", http.StatusInternalServerError)
		return
	}

	writeJSON(w, requests, http.StatusOK)
}

// createServiceRequest handles POST /api/service-requests
func createServiceRequest(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User) {
	var input struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Category    string  `json:"category"`
		CommunityID uint    `json:"community_id"`
		Budget      float64 `json:"budget"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.Title == "" || input.Description == "" || input.CommunityID == 0 {
		writeError(w, "Title, description, and community_id are required", http.StatusBadRequest)
		return
	}

	// Create service request
	request := ServiceRequest{
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		RequesterID: user.ID,
		CommunityID: input.CommunityID,
		Status:      "open",
		Budget:      input.Budget,
	}

	if err := db.Create(&request).Error; err != nil {
		writeError(w, "Failed to create service request", http.StatusInternalServerError)
		return
	}

	// Reload with associations
	if err := db.Preload("Requester").Preload("Community").First(&request, request.ID).Error; err != nil {
		writeError(w, "Failed to load created request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, request, http.StatusCreated)
}

// serviceRequestDetailHandler handles GET, PUT, DELETE for a specific service request
func serviceRequestDetailHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromContext(r, db)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/service-requests/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeError(w, "Service request ID required", http.StatusBadRequest)
			return
		}

		requestID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid service request ID", http.StatusBadRequest)
			return
		}

		// Handle accept offer endpoint
		if len(parts) >= 2 && parts[1] == "accept-offer" {
			acceptServiceOffer(w, r, db, user, uint(requestID))
			return
		}

		switch r.Method {
		case http.MethodGet:
			getServiceRequest(w, r, db, uint(requestID))
		case http.MethodPut:
			updateServiceRequest(w, r, db, user, uint(requestID))
		case http.MethodDelete:
			deleteServiceRequest(w, r, db, user, uint(requestID))
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// getServiceRequest handles GET /api/service-requests/:id
func getServiceRequest(w http.ResponseWriter, r *http.Request, db *gorm.DB, requestID uint) {
	var request ServiceRequest
	if err := db.Preload("Requester").
		Preload("Community").
		Preload("ServiceOffers").
		Preload("ServiceOffers.Provider").
		Preload("AcceptedOffer").
		Preload("AcceptedOffer.Provider").
		First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service request not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service request", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, request, http.StatusOK)
}

// updateServiceRequest handles PUT /api/service-requests/:id
func updateServiceRequest(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, requestID uint) {
	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service request not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service request", http.StatusInternalServerError)
		}
		return
	}

	// Only requester can update
	if request.RequesterID != user.ID && user.Role != RoleSuperAdmin && user.Role != RoleAdmin {
		writeError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var input struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Category    *string  `json:"category"`
		Budget      *float64 `json:"budget"`
		Status      *string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
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
		writeError(w, "Failed to update service request", http.StatusInternalServerError)
		return
	}

	// Reload with associations
	if err := db.Preload("Requester").Preload("Community").First(&request, requestID).Error; err != nil {
		writeError(w, "Failed to load updated request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, request, http.StatusOK)
}

// deleteServiceRequest handles DELETE /api/service-requests/:id
func deleteServiceRequest(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, requestID uint) {
	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service request not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service request", http.StatusInternalServerError)
		}
		return
	}

	// Only requester or admin can delete
	if request.RequesterID != user.ID && user.Role != RoleSuperAdmin && user.Role != RoleAdmin {
		writeError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Soft delete
	if err := db.Delete(&request).Error; err != nil {
		writeError(w, "Failed to delete service request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// acceptServiceOffer handles PUT /api/service-requests/:id/accept-offer
func acceptServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, requestID uint) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		OfferID uint `json:"offer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.OfferID == 0 {
		writeError(w, "offer_id is required", http.StatusBadRequest)
		return
	}

	// Verify request exists and user is requester
	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service request not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service request", http.StatusInternalServerError)
		}
		return
	}

	if request.RequesterID != user.ID {
		writeError(w, "Unauthorized: only requester can accept offers", http.StatusForbidden)
		return
	}

	// Verify offer exists and belongs to this request
	var offer ServiceOffer
	if err := db.First(&offer, input.OfferID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Offer not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch offer", http.StatusInternalServerError)
		}
		return
	}

	if offer.ServiceRequestID != requestID {
		writeError(w, "Offer does not belong to this request", http.StatusBadRequest)
		return
	}

	// Update request and offer in transaction
	err := db.Transaction(func(tx *gorm.DB) error {
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
		writeError(w, "Failed to accept offer", http.StatusInternalServerError)
		return
	}

	// Reload request with associations
	if err := db.Preload("Requester").
		Preload("Community").
		Preload("AcceptedOffer").
		Preload("AcceptedOffer.Provider").
		First(&request, requestID).Error; err != nil {
		writeError(w, "Failed to load updated request", http.StatusInternalServerError)
		return
	}

	writeJSON(w, request, http.StatusOK)
}

// serviceOffersHandler handles GET (list) and POST (create) for service offers
func serviceOffersHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromContext(r, db)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			listServiceOffers(w, r, db, user)
		case http.MethodPost:
			createServiceOffer(w, r, db, user)
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// listServiceOffers handles GET /api/service-offers
func listServiceOffers(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User) {
	// Get query parameters
	serviceRequestIDStr := r.URL.Query().Get("service_request_id")
	myOffers := r.URL.Query().Get("my_offers") == "true"
	providerIDStr := r.URL.Query().Get("provider_id")

	query := db.Model(&ServiceOffer{}).
		Preload("Provider").
		Preload("ServiceRequest").
		Preload("ServiceRequest.Requester").
		Preload("ServiceRequest.Community")

	// Filter by service request if specified
	if serviceRequestIDStr != "" {
		serviceRequestID, err := strconv.ParseUint(serviceRequestIDStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid service_request_id", http.StatusBadRequest)
			return
		}
		query = query.Where("service_request_id = ?", serviceRequestID)
	}

	// Filter by current user's offers if requested
	if myOffers {
		query = query.Where("provider_id = ?", user.ID)
	}

	// Filter by provider ID if specified
	if providerIDStr != "" {
		providerID, err := strconv.ParseUint(providerIDStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid provider_id", http.StatusBadRequest)
			return
		}
		query = query.Where("provider_id = ?", providerID)
	}

	var offers []ServiceOffer
	if err := query.Order("created_at DESC").Find(&offers).Error; err != nil {
		writeError(w, "Failed to fetch service offers", http.StatusInternalServerError)
		return
	}

	writeJSON(w, offers, http.StatusOK)
}

// createServiceOffer handles POST /api/service-offers
func createServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User) {
	var input struct {
		ServiceRequestID  uint    `json:"service_request_id"`
		Description       string  `json:"description"`
		ProposedPrice     float64 `json:"proposed_price"`
		EstimatedDuration string  `json:"estimated_duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.ServiceRequestID == 0 || input.Description == "" {
		writeError(w, "service_request_id and description are required", http.StatusBadRequest)
		return
	}

	// Verify service request exists and is open
	var request ServiceRequest
	if err := db.First(&request, input.ServiceRequestID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service request not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service request", http.StatusInternalServerError)
		}
		return
	}

	if request.Status != "open" {
		writeError(w, "Cannot create offer for non-open requests", http.StatusBadRequest)
		return
	}

	// Create service offer
	offer := ServiceOffer{
		ServiceRequestID:  input.ServiceRequestID,
		ProviderID:        user.ID,
		Description:       input.Description,
		ProposedPrice:     input.ProposedPrice,
		EstimatedDuration: input.EstimatedDuration,
		Status:            "pending",
	}

	if err := db.Create(&offer).Error; err != nil {
		writeError(w, "Failed to create service offer", http.StatusInternalServerError)
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offer.ID).Error; err != nil {
		writeError(w, "Failed to load created offer", http.StatusInternalServerError)
		return
	}

	writeJSON(w, offer, http.StatusCreated)
}

// serviceOfferDetailHandler handles GET, PUT, DELETE for a specific service offer
func serviceOfferDetailHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserFromContext(r, db)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/service-offers/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeError(w, "Service offer ID required", http.StatusBadRequest)
			return
		}

		offerID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid service offer ID", http.StatusBadRequest)
			return
		}

		// Handle withdraw endpoint
		if len(parts) >= 2 && parts[1] == "withdraw" {
			withdrawServiceOffer(w, r, db, user, uint(offerID))
			return
		}

		switch r.Method {
		case http.MethodGet:
			getServiceOffer(w, r, db, uint(offerID))
		case http.MethodPut:
			updateServiceOffer(w, r, db, user, uint(offerID))
		case http.MethodDelete:
			deleteServiceOffer(w, r, db, user, uint(offerID))
		default:
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// withdrawServiceOffer handles POST /api/service-offers/:id/withdraw
func withdrawServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, offerID uint) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service offer not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service offer", http.StatusInternalServerError)
		}
		return
	}

	// Only provider can withdraw
	if offer.ProviderID != user.ID {
		writeError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Cannot withdraw accepted offers
	if offer.Status == "accepted" {
		writeError(w, "Cannot withdraw accepted offers", http.StatusBadRequest)
		return
	}

	// Update status to withdrawn
	if err := db.Model(&offer).Update("status", "withdrawn").Error; err != nil {
		writeError(w, "Failed to withdraw service offer", http.StatusInternalServerError)
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offerID).Error; err != nil {
		writeError(w, "Failed to load updated offer", http.StatusInternalServerError)
		return
	}

	writeJSON(w, offer, http.StatusOK)
}

// getServiceOffer handles GET /api/service-offers/:id
func getServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, offerID uint) {
	var offer ServiceOffer
	if err := db.Preload("Provider").
		Preload("ServiceRequest").
		Preload("ServiceRequest.Requester").
		First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service offer not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service offer", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, offer, http.StatusOK)
}

// updateServiceOffer handles PUT /api/service-offers/:id
func updateServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, offerID uint) {
	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service offer not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service offer", http.StatusInternalServerError)
		}
		return
	}

	// Only provider can update
	if offer.ProviderID != user.ID {
		writeError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var input struct {
		Description       *string  `json:"description"`
		ProposedPrice     *float64 `json:"proposed_price"`
		EstimatedDuration *string  `json:"estimated_duration"`
		Status            *string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
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
		writeError(w, "Failed to update service offer", http.StatusInternalServerError)
		return
	}

	// Reload with associations
	if err := db.Preload("Provider").Preload("ServiceRequest").First(&offer, offerID).Error; err != nil {
		writeError(w, "Failed to load updated offer", http.StatusInternalServerError)
		return
	}

	writeJSON(w, offer, http.StatusOK)
}

// deleteServiceOffer handles DELETE /api/service-offers/:id (withdraw offer)
func deleteServiceOffer(w http.ResponseWriter, r *http.Request, db *gorm.DB, user *User, offerID uint) {
	var offer ServiceOffer
	if err := db.First(&offer, offerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			writeError(w, "Service offer not found", http.StatusNotFound)
		} else {
			writeError(w, "Failed to fetch service offer", http.StatusInternalServerError)
		}
		return
	}

	// Only provider can delete
	if offer.ProviderID != user.ID {
		writeError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Cannot delete accepted offers
	if offer.Status == "accepted" {
		writeError(w, "Cannot withdraw accepted offers", http.StatusBadRequest)
		return
	}

	// Soft delete
	if err := db.Delete(&offer).Error; err != nil {
		writeError(w, "Failed to delete service offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
