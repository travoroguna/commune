package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-secret-change-in-production")
	}
}

type Claims struct {
	UserID uint     `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(user *User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("MODE") == "production",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("MODE") == "production",
		MaxAge:   -1,
	})
}

func getAuthToken(r *http.Request) string {
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value
	}
	return ""
}

// Middleware to authenticate requests
func authMiddleware(db *gorm.DB) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := getAuthToken(r)
			if token == "" {
				writeError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := validateToken(token)
			if err != nil {
				writeError(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			var user User
			if err := db.First(&user, claims.UserID).Error; err != nil {
				writeError(w, "User not found", http.StatusUnauthorized)
				return
			}

			if !user.IsActive {
				writeError(w, "User is inactive", http.StatusForbidden)
				return
			}

			// Store user ID in context-like manner (using request header for simplicity)
			r.Header.Set("X-User-ID", strconv.FormatUint(uint64(user.ID), 10))
			r.Header.Set("X-User-Role", string(user.Role))

			next(w, r)
		}
	}
}

// Middleware to require specific roles
func requireRole(db *gorm.DB, roles ...UserRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
			userRole := UserRole(r.Header.Get("X-User-Role"))
			
			allowed := false
			for _, role := range roles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				writeError(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}

func getCurrentUser(r *http.Request) (uint, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, errors.New("user not authenticated")
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	return uint(userID), err
}

// Auth handlers

func loginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" {
			writeError(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		var user User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			writeError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !user.IsActive {
			writeError(w, "User is inactive", http.StatusForbidden)
			return
		}

		if !checkPasswordHash(req.Password, user.PasswordHash) {
			writeError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := generateToken(&user)
		if err != nil {
			writeError(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		setAuthCookie(w, token)

		writeJSON(w, map[string]interface{}{
			"user":  sanitizeUser(&user),
			"token": token,
		}, http.StatusOK)
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		clearAuthCookie(w)
		writeJSON(w, map[string]interface{}{"message": "Logged out successfully"}, http.StatusOK)
	}
}

func getCurrentUserHandler(db *gorm.DB) http.HandlerFunc {
	return authMiddleware(db)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, err := getCurrentUser(r)
		if err != nil {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			writeError(w, "User not found", http.StatusNotFound)
			return
		}

		writeJSON(w, sanitizeUser(&user), http.StatusOK)
	})
}

func checkFirstBootHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var count int64
		db.Model(&User{}).Count(&count)

		writeJSON(w, map[string]bool{"needsSetup": count == 0}, http.StatusOK)
	}
}

func setupSuperUserHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var count int64
		db.Model(&User{}).Count(&count)
		if count > 0 {
			writeError(w, "Super user already exists", http.StatusForbidden)
			return
		}

		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			writeError(w, "Name, email and password are required", http.StatusBadRequest)
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
			Role:         RoleSuperAdmin,
			IsActive:     true,
		}

		if err := db.Create(&user).Error; err != nil {
			writeError(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		token, err := generateToken(&user)
		if err != nil {
			writeError(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		setAuthCookie(w, token)

		writeJSON(w, map[string]interface{}{
			"user":  sanitizeUser(&user),
			"token": token,
		}, http.StatusCreated)
	}
}

func sanitizeUser(user *User) map[string]interface{} {
	return map[string]interface{}{
		"ID":        user.ID,
		"CreatedAt": user.CreatedAt,
		"UpdatedAt": user.UpdatedAt,
		"DeletedAt": user.DeletedAt,
		"Name":      user.Name,
		"Email":     user.Email,
		"Role":      user.Role,
		"IsActive":  user.IsActive,
	}
}
