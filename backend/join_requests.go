package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Join Request handlers

func getJoinRequestsHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var joinRequests []JoinRequest
		if err := db.Preload("User").Preload("Community").Where("status = ?", "pending").Find(&joinRequests).Error; err != nil {
			writeError(w, "Failed to fetch join requests", http.StatusInternalServerError)
			return
		}

		writeJSON(w, joinRequests, http.StatusOK)
	})
}

func getCommunityJoinRequestsHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/communities/"), "/")
		if len(parts) < 2 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		communityID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		var joinRequests []JoinRequest
		if err := db.Preload("User").Preload("Community").Where("community_id = ? AND status = ?", communityID, "pending").Find(&joinRequests).Error; err != nil {
			writeError(w, "Failed to fetch join requests", http.StatusInternalServerError)
			return
		}

		writeJSON(w, joinRequests, http.StatusOK)
	})
}

func createJoinRequestHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			CommunityID uint   `json:"communityId"`
			Message     string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.CommunityID == 0 {
			writeError(w, "Community ID is required", http.StatusBadRequest)
			return
		}

		// Check if community exists
		var community Community
		if err := db.First(&community, req.CommunityID).Error; err != nil {
			writeError(w, "Community not found", http.StatusNotFound)
			return
		}

		// Check if user is already a member
		var existing UserCommunity
		err = db.Where("user_id = ? AND community_id = ?", userID, req.CommunityID).First(&existing).Error
		if err == nil {
			writeError(w, "You are already a member of this community", http.StatusConflict)
			return
		}

		// Check if there's already a pending request
		var existingRequest JoinRequest
		err = db.Where("user_id = ? AND community_id = ? AND status = ?", userID, req.CommunityID, "pending").First(&existingRequest).Error
		if err == nil {
			writeError(w, "You already have a pending request for this community", http.StatusConflict)
			return
		}

		joinRequest := JoinRequest{
			UserID:      userID,
			CommunityID: req.CommunityID,
			Status:      "pending",
			Message:     req.Message,
		}

		if err := db.Create(&joinRequest).Error; err != nil {
			writeError(w, "Failed to create join request", http.StatusInternalServerError)
			return
		}

		// Preload relationships
		db.Preload("User").Preload("Community").First(&joinRequest, joinRequest.ID)

		writeJSON(w, joinRequest, http.StatusCreated)
	})
}

func approveJoinRequestHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/join-requests/"), "/")
		if len(parts) < 2 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		requestID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid request ID", http.StatusBadRequest)
			return
		}

		var req struct {
			Role UserRole `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Role == "" {
			req.Role = RoleUser
		}

		var joinRequest JoinRequest
		if err := db.First(&joinRequest, requestID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Join request not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch join request", http.StatusInternalServerError)
			}
			return
		}

		if joinRequest.Status != "pending" {
			writeError(w, "This request has already been processed", http.StatusBadRequest)
			return
		}

		// Update request status
		joinRequest.Status = "approved"
		if err := db.Save(&joinRequest).Error; err != nil {
			writeError(w, "Failed to update join request", http.StatusInternalServerError)
			return
		}

		// Add user to community
		userCommunity := UserCommunity{
			UserID:      joinRequest.UserID,
			CommunityID: joinRequest.CommunityID,
			Role:        req.Role,
			IsActive:    true,
		}

		if err := db.Create(&userCommunity).Error; err != nil {
			// If adding member fails, revert the join request status
			joinRequest.Status = "pending"
			db.Save(&joinRequest)
			writeError(w, "Failed to add user to community", http.StatusInternalServerError)
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").First(&joinRequest, requestID)

		writeJSON(w, joinRequest, http.StatusOK)
	})
}

func rejectJoinRequestHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/join-requests/"), "/")
		if len(parts) < 2 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		requestID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid request ID", http.StatusBadRequest)
			return
		}

		var joinRequest JoinRequest
		if err := db.First(&joinRequest, requestID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Join request not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch join request", http.StatusInternalServerError)
			}
			return
		}

		if joinRequest.Status != "pending" {
			writeError(w, "This request has already been processed", http.StatusBadRequest)
			return
		}

		joinRequest.Status = "rejected"
		if err := db.Save(&joinRequest).Error; err != nil {
			writeError(w, "Failed to update join request", http.StatusInternalServerError)
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").First(&joinRequest, requestID)

		writeJSON(w, joinRequest, http.StatusOK)
	})
}
