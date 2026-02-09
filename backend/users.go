package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// User handlers

func getUsersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		var users []User
		if err := db.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch users"})
			return
		}

		result := make([]map[string]interface{}, len(users))
		for i, user := range users {
			result[i] = sanitizeUser(&user)
		}

		c.JSON(http.StatusOK, result)
	}
}

func getUserByIDHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user"})
			}
			return
		}

		c.JSON(http.StatusOK, sanitizeUser(&user))
	}
}

func createUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		var req struct {
			Name     string   `json:"name"`
			Email    string   `json:"email"`
			Password string   `json:"password"`
			Role     UserRole `json:"role"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Name, email and password are required"})
			return
		}

		if req.Role == "" {
			req.Role = RoleUser
		}

		// Check if user already exists
		var existingUser User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"message": "User with this email already exists"})
			return
		}

		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
			return
		}

		user := User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: passwordHash,
			Role:         req.Role,
			IsActive:     true,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, sanitizeUser(&user))
	}
}

func updateUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		currentUserID, err := getCurrentUser(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user"})
			}
			return
		}

		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		currentUserRole, _ := c.Get("userRole")

		// Check permissions
		if uint(id) != currentUserID && currentUserRole != RoleSuperAdmin && currentUserRole != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient permissions"})
			return
		}

		// Only admins can change roles
		if _, hasRole := req["Role"]; hasRole && currentUserRole != RoleSuperAdmin && currentUserRole != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"message": "Only admins can change user roles"})
			return
		}

		updates := make(map[string]interface{})
		if name, ok := req["Name"].(string); ok && name != "" {
			updates["name"] = name
		}
		if email, ok := req["Email"].(string); ok && email != "" {
			updates["email"] = email
		}
		if role, ok := req["Role"].(string); ok && role != "" {
			updates["role"] = role
		}
		if isActive, ok := req["IsActive"].(bool); ok {
			updates["is_active"] = isActive
		}

		if len(updates) > 0 {
			if err := db.Model(&user).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user"})
				return
			}
		}

		// Fetch updated user
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch updated user"})
			return
		}

		c.JSON(http.StatusOK, sanitizeUser(&user))
	}
}

func deleteUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		requireRole(db, RoleSuperAdmin, RoleAdmin)(c)
		if c.IsAborted() {
			return
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user"})
			}
			return
		}

		// Soft delete
		if err := db.Delete(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func changePasswordHandler(db *gorm.DB) gin.HandlerFunc {
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
			OldPassword string `json:"oldPassword"`
			NewPassword string `json:"newPassword"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.OldPassword == "" || req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Old password and new password are required"})
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		if !checkPasswordHash(req.OldPassword, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Old password is incorrect"})
			return
		}

		passwordHash, err := hashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to hash password"})
			return
		}

		if err := db.Model(&user).Update("password_hash", passwordHash).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
	}
}

func getUserCommunitiesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
			return
		}

		var userCommunities []UserCommunity
		if err := db.Preload("Community").Where("user_id = ? AND is_active = ?", id, true).Find(&userCommunities).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch user communities"})
			return
		}

		c.JSON(http.StatusOK, userCommunities)
	}
}
