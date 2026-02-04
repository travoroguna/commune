package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// User handlers

func getUsersHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var users []User
		if err := db.Where("deleted_at IS NULL").Find(&users).Error; err != nil {
			writeError(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}

		result := make([]map[string]interface{}, len(users))
		for i, user := range users {
			result[i] = sanitizeUser(&user)
		}

		writeJSON(w, result, http.StatusOK)
	})
}

func getUserByIDHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "User not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch user", http.StatusInternalServerError)
			}
			return
		}

		writeJSON(w, sanitizeUser(&user), http.StatusOK)
	})
}

func createUserHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Name     string   `json:"name"`
			Email    string   `json:"email"`
			Password string   `json:"password"`
			Role     UserRole `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			writeError(w, "Name, email and password are required", http.StatusBadRequest)
			return
		}

		if req.Role == "" {
			req.Role = RoleUser
		}

		// Check if user already exists
		var existingUser User
		if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			writeError(w, "User with this email already exists", http.StatusConflict)
			return
		}

		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			writeError(w, "Failed to hash password", http.StatusInternalServerError)
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
			writeError(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		writeJSON(w, sanitizeUser(&user), http.StatusCreated)
	})
}

func updateUserHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		currentUserID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "User not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch user", http.StatusInternalServerError)
			}
			return
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		currentUserRole := UserRole(r.Header.Get("X-User-Role"))

		// Check permissions
		if uint(id) != currentUserID && currentUserRole != RoleSuperAdmin && currentUserRole != RoleAdmin {
			writeError(w, "Insufficient permissions", http.StatusForbidden)
			return
		}

		// Only admins can change roles
		if _, hasRole := req["Role"]; hasRole && currentUserRole != RoleSuperAdmin && currentUserRole != RoleAdmin {
			writeError(w, "Only admins can change user roles", http.StatusForbidden)
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
				writeError(w, "Failed to update user", http.StatusInternalServerError)
				return
			}
		}

		// Fetch updated user
		if err := db.First(&user, id).Error; err != nil {
			writeError(w, "Failed to fetch updated user", http.StatusInternalServerError)
			return
		}

		writeJSON(w, sanitizeUser(&user), http.StatusOK)
	})
}

func deleteUserHandler(db *gorm.DB) http.HandlerFunc {
	return requireRole(db, RoleSuperAdmin, RoleAdmin)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var user User
		if err := db.First(&user, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				writeError(w, "User not found", http.StatusNotFound)
			} else {
				writeError(w, "Failed to fetch user", http.StatusInternalServerError)
			}
			return
		}

		// Soft delete
		if err := db.Delete(&user).Error; err != nil {
			writeError(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{"message": "User deleted successfully"}, http.StatusOK)
	})
}

func changePasswordHandler(db *gorm.DB) http.HandlerFunc {
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
			OldPassword string `json:"oldPassword"`
			NewPassword string `json:"newPassword"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.OldPassword == "" || req.NewPassword == "" {
			writeError(w, "Old password and new password are required", http.StatusBadRequest)
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			writeError(w, "User not found", http.StatusNotFound)
			return
		}

		if !checkPasswordHash(req.OldPassword, user.PasswordHash) {
			writeError(w, "Old password is incorrect", http.StatusUnauthorized)
			return
		}

		passwordHash, err := hashPassword(req.NewPassword)
		if err != nil {
			writeError(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		if err := db.Model(&user).Update("password_hash", passwordHash).Error; err != nil {
			writeError(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		writeJSON(w, map[string]interface{}{"message": "Password changed successfully"}, http.StatusOK)
	})
}

func getUserCommunitiesHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/users/"), "/")
		if len(parts) < 2 {
			writeError(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		id, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			writeError(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var userCommunities []UserCommunity
		if err := db.Preload("Community").Where("user_id = ? AND is_active = ?", id, true).Find(&userCommunities).Error; err != nil {
			writeError(w, "Failed to fetch user communities", http.StatusInternalServerError)
			return
		}

		writeJSON(w, userCommunities, http.StatusOK)
	})
}
