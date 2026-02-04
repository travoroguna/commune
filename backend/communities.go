package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Community handlers

func getCommunitiesHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var communities []Community
		if err := db.Where("is_active = ?", true).Find(&communities).Error; err != nil {
			writeError(w, "Failed to fetch communities", http.StatusInternalServerError)
			return
		}

		writeJSON(w, communities, http.StatusOK)
	})
}

func getCommunityByIDHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/communities/")
		idStr = strings.Split(idStr, "/")[0]
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Community not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch community", http.StatusInternalServerError)
			}
			return
		}

		writeJSON(w, community, http.StatusOK)
	})
}

func createCommunityHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Community
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			writeError(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Generate slug from name if not provided
		if req.Slug == "" {
			req.Slug = GenerateSlug(req.Name)
		} else {
			req.Slug = GenerateSlug(req.Slug)
		}

		// Check if slug already exists
		var existingCommunity Community
		if err := db.Where("slug = ?", req.Slug).First(&existingCommunity).Error; err == nil {
			writeError(w, "Community with this slug already exists", http.StatusConflict)
			return
		}

		req.IsActive = true

		if err := db.Create(&req).Error; err != nil {
			writeError(w, "Failed to create community", http.StatusInternalServerError)
			return
		}

		writeJSON(w, req, http.StatusCreated)
	})
}

func updateCommunityHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/communities/")
		idStr = strings.Split(idStr, "/")[0]
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Community not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch community", http.StatusInternalServerError)
			}
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updates := make(map[string]interface{})
		if name, ok := req["Name"].(string); ok && name != "" {
			updates["name"] = name
			// Regenerate slug if name changes
			if _, hasSlug := req["Slug"]; !hasSlug {
				updates["slug"] = GenerateSlug(name)
			}
		}
		if slug, ok := req["Slug"].(string); ok && slug != "" {
			updates["slug"] = GenerateSlug(slug)
		}
		if description, ok := req["Description"].(string); ok {
			updates["description"] = description
		}
		if subdomain, ok := req["Subdomain"].(string); ok {
			updates["subdomain"] = subdomain
		}
		if customDomain, ok := req["CustomDomain"].(string); ok {
			updates["custom_domain"] = customDomain
		}
		if address, ok := req["Address"].(string); ok {
			updates["address"] = address
		}
		if city, ok := req["City"].(string); ok {
			updates["city"] = city
		}
		if state, ok := req["State"].(string); ok {
			updates["state"] = state
		}
		if country, ok := req["Country"].(string); ok {
			updates["country"] = country
		}
		if zipCode, ok := req["ZipCode"].(string); ok {
			updates["zip_code"] = zipCode
		}
		if isActive, ok := req["IsActive"].(bool); ok {
			updates["is_active"] = isActive
		}

		if len(updates) > 0 {
			if err := db.Model(&community).Updates(updates).Error; err != nil {
				writeError(w, "Failed to update community", http.StatusInternalServerError)
				return
			}
		}

		// Fetch updated community
		if err := db.First(&community, id).Error; err != nil {
			writeError(w, "Failed to fetch updated community", http.StatusInternalServerError)
			return
		}

		writeJSON(w, community, http.StatusOK)
	})
}

func deleteCommunityHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/communities/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Community not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch community", http.StatusInternalServerError)
			}
			return
		}

		// Soft delete
		if err := db.Delete(&community).Error; err != nil {
			writeError(w, "Failed to delete community", http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{"message": "Community deleted successfully"}, http.StatusOK)
	})
}

func getCommunityMembersHandler(db *gorm.DB) http.HandlerFunc {
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

		id, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		var userCommunities []UserCommunity
		if err := db.Preload("User").Where("community_id = ? AND is_active = ?", id, true).Find(&userCommunities).Error; err != nil {
			writeError(w, "Failed to fetch community members", http.StatusInternalServerError)
			return
		}

		writeJSON(w, userCommunities, http.StatusOK)
	})
}

func addCommunityMemberHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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

		var req struct {
			UserID uint     `json:"userId"`
			Role   UserRole `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.UserID == 0 {
			writeError(w, "User ID is required", http.StatusBadRequest)
			return
		}

		if req.Role == "" {
			req.Role = RoleUser
		}

		// Check if user exists
		var user User
		if err := db.First(&user, req.UserID).Error; err != nil {
			writeError(w, "User not found", http.StatusNotFound)
			return
		}

		// Check if community exists
		var community Community
		if err := db.First(&community, communityID).Error; err != nil {
			writeError(w, "Community not found", http.StatusNotFound)
			return
		}

		// Check if membership already exists
		var existing UserCommunity
		err = db.Where("user_id = ? AND community_id = ?", req.UserID, communityID).First(&existing).Error
		if err == nil {
			writeError(w, "User is already a member of this community", http.StatusConflict)
			return
		}

		userCommunity := UserCommunity{
			UserID:      req.UserID,
			CommunityID: uint(communityID),
			Role:        req.Role,
			IsActive:    true,
		}

		if err := db.Create(&userCommunity).Error; err != nil {
			writeError(w, "Failed to add member", http.StatusInternalServerError)
			return
		}

		// Preload relationships
		db.Preload("User").Preload("Community").First(&userCommunity, "user_id = ? AND community_id = ?", req.UserID, communityID)

		writeJSON(w, userCommunity, http.StatusCreated)
	})
}

func removeCommunityMemberHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/communities/"), "/")
		if len(parts) < 3 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		communityID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		userID, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		result := db.Where("user_id = ? AND community_id = ?", userID, communityID).Delete(&UserCommunity{})
		if result.Error != nil {
			writeError(w, "Failed to remove member", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			writeError(w, "Member not found", http.StatusNotFound)
			return
		}

		writeJSON(w, map[string]interface{}{"message": "Member removed successfully"}, http.StatusOK)
	})
}

func updateCommunityMemberRoleHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/communities/"), "/")
		if len(parts) < 3 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		communityID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid community ID", http.StatusBadRequest)
			return
		}

		userID, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
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
			writeError(w, "Role is required", http.StatusBadRequest)
			return
		}

		var userCommunity UserCommunity
		if err := db.Where("user_id = ? AND community_id = ?", userID, communityID).First(&userCommunity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "Member not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch member", http.StatusInternalServerError)
			}
			return
		}

		if err := db.Model(&userCommunity).Update("role", req.Role).Error; err != nil {
			writeError(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").Where("user_id = ? AND community_id = ?", userID, communityID).First(&userCommunity)

		writeJSON(w, userCommunity, http.StatusOK)
	})
}
