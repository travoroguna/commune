package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Join Request handlers

func getJoinRequestsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		var joinRequests []JoinRequest
		if err := db.Preload("User").Preload("Community").Where("status = ?", "pending").Find(&joinRequests).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch join requests"})
			return
		}

		c.JSON(http.StatusOK, joinRequests)
	}
}

func getCommunityJoinRequestsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		communityID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid community ID"})
			return
		}

		var joinRequests []JoinRequest
		if err := db.Preload("User").Preload("Community").Where("community_id = ? AND status = ?", communityID, "pending").Find(&joinRequests).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch join requests"})
			return
		}

		c.JSON(http.StatusOK, joinRequests)
	}
}

func createJoinRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		userID, err := getCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		var req struct {
			CommunityID uint   `json:"communityId"`
			Message     string `json:"message"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.CommunityID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Community ID is required"})
			return
		}

		// Check if community exists
		var community Community
		if err := db.First(&community, req.CommunityID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Community not found"})
			return
		}

		// Check if user is already a member
		var existing UserCommunity
		err = db.Where("user_id = ? AND community_id = ?", userID, req.CommunityID).First(&existing).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "You are already a member of this community"})
			return
		}

		// Check if there's already a pending request
		var existingRequest JoinRequest
		err = db.Where("user_id = ? AND community_id = ? AND status = ?", userID, req.CommunityID, "pending").First(&existingRequest).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "You already have a pending request for this community"})
			return
		}

		joinRequest := JoinRequest{
			UserID:      userID,
			CommunityID: req.CommunityID,
			Status:      "pending",
			Message:     req.Message,
		}

		if err := db.Create(&joinRequest).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create join request"})
			return
		}

		// Preload relationships
		db.Preload("User").Preload("Community").First(&joinRequest, joinRequest.ID)

		c.JSON(http.StatusCreated, joinRequest)
	}
}

func approveJoinRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		requestID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request ID"})
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
			req.Role = RoleUser
		}

		var joinRequest JoinRequest
		if err := db.First(&joinRequest, requestID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Join request not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch join request"})
			}
			return
		}

		if joinRequest.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "This request has already been processed"})
			return
		}

		// Update request status
		joinRequest.Status = "approved"
		if err := db.Save(&joinRequest).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update join request"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to add user to community"})
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").First(&joinRequest, requestID)

		c.JSON(http.StatusOK, joinRequest)
	}
}

func rejectJoinRequestHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		idStr := c.Param("id")
		requestID, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request ID"})
			return
		}

		var joinRequest JoinRequest
		if err := db.First(&joinRequest, requestID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "Join request not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch join request"})
			}
			return
		}

		if joinRequest.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "This request has already been processed"})
			return
		}

		joinRequest.Status = "rejected"
		if err := db.Save(&joinRequest).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update join request"})
			return
		}

		// Reload with relationships
		db.Preload("User").Preload("Community").First(&joinRequest, requestID)

		c.JSON(http.StatusOK, joinRequest)
	}
}
