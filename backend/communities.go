package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Community handlers

func getCommunitiesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		var communities []Community
		if err := db.Where("is_active = ?", true).Find(&communities).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch communities"})
			return
		}

		c.JSON(http.StatusOK, communities)
	}
}

func getCommunityByIDHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Community not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch community"})
			}
			return
		}

		c.JSON(http.StatusOK, community)
	}
}

func createCommunityHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin)(c)
		if c.IsAborted() {
			return
		}

		var req Community
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Name is required"})
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
			c.JSON(http.StatusConflict, gin.H{"message": "Community with this slug already exists"})
			return
		}

		req.IsActive = true

		if err := db.Create(&req).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create community"})
			return
		}

		c.JSON(http.StatusCreated, req)
	}
}

func updateCommunityHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Community not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch community"})
			}
			return
		}

		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
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
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update community"})
				return
			}
		}

		// Fetch updated community
		if err := db.First(&community, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch updated community"})
			return
		}

		c.JSON(http.StatusOK, community)
	}
}

func deleteCommunityHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var community Community
		if err := db.First(&community, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Community not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch community"})
			}
			return
		}

		// Soft delete
		if err := db.Delete(&community).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete community"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Community deleted successfully"})
	}
}

func getCommunityMembersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var userCommunities []UserCommunity
		if err := db.Preload("User").Where("community_id = ? AND is_active = ?", id, true).Find(&userCommunities).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch community members"})
			return
		}

		c.JSON(http.StatusOK, userCommunities)
	}
}

func addCommunityMemberHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		communityID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var req struct {
			UserID uint     `json:"userId"`
			Role   UserRole `json:"role"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.UserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "User ID is required"})
			return
		}

		if req.Role == "" {
			req.Role = RoleUser
		}

		// Check if user exists
		var user User
		if err := db.First(&user, req.UserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		// Check if community exists
		var community Community
		if err := db.First(&community, communityID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Community not found"})
			return
		}

		// Check if membership already exists
		var existing UserCommunity
		err = db.Where("user_id = ? AND community_id = ?", req.UserID, communityID).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "User is already a member of this community"})
			return
		}

		userCommunity := UserCommunity{
			UserID:      req.UserID,
			CommunityID: uint(communityID),
			Role:        req.Role,
			IsActive:    true,
		}

		if err := db.Create(&userCommunity).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add member"})
			return
		}

		// Preload relationships
		db.Preload("User").Preload("Community").First(&userCommunity, "user_id = ? AND community_id = ?", req.UserID, communityID)

		c.JSON(http.StatusCreated, userCommunity)
	}
}

func removeCommunityMemberHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		communityIDStr := c.Param("id")
		communityID, err := strconv.ParseUint(communityIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		userIDStr := c.Param("userID")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		result := db.Where("user_id = ? AND community_id = ?", userID, communityID).Delete(&UserCommunity{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to remove member"})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Member not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
	}
}

func updateCommunityMemberRoleHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		communityIDStr := c.Param("id")
		communityID, err := strconv.ParseUint(communityIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		userIDStr := c.Param("userID")
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		var req struct {
			Role UserRole `json:"role"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Role is required"})
			return
		}

		var userCommunity UserCommunity
		if err := db.Where("user_id = ? AND community_id = ?", userID, communityID).First(&userCommunity).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Member not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch member"})
			}
			return
		}

		if err := db.Model(&userCommunity).Update("role", req.Role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update member role"})
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").Where("user_id = ? AND community_id = ?", userID, communityID).First(&userCommunity)

		c.JSON(http.StatusOK, userCommunity)
	}
}
