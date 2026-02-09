package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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

func setAuthCookie(c *gin.Context, token string) {
	c.SetCookie(
		"auth_token",
		token,
		86400, // 24 hours
		"/",
		"",
		os.Getenv("MODE") == "production",
		true, // HttpOnly
	)
}

func clearAuthCookie(c *gin.Context) {
	c.SetCookie(
		"auth_token",
		"",
		-1,
		"/",
		"",
		os.Getenv("MODE") == "production",
		true, // HttpOnly
	)
}

func getAuthToken(c *gin.Context) string {
	token, err := c.Cookie("auth_token")
	if err == nil {
		return token
	}
	return ""
}

// Middleware to authenticate requests
func authMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := getAuthToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		claims, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			c.Abort()
			return
		}

		var user User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
			c.Abort()
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"message": "User is inactive"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.Set("user", &user)

		c.Next()
	}
}

// Middleware to require specific roles
func requireRole(db *gorm.DB, roles ...UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authMiddleware(db)(c)
		if c.IsAborted() {
			return
		}

		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient permissions"})
			c.Abort()
			return
		}

		allowed := false
		for _, role := range roles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getCurrentUser(c *gin.Context) (uint, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("user not authenticated")
	}
	return userID.(uint), nil
}

// Auth handlers

func loginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Email and password are required"})
			return
		}

		var user User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
			return
		}

		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"message": "User is inactive"})
			return
		}

		if !checkPasswordHash(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
			return
		}

		token, err := generateToken(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
			return
		}

		setAuthCookie(c, token)

		c.JSON(http.StatusOK, gin.H{
			"user":  sanitizeUser(&user),
			"token": token,
		})
	}
}

func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		clearAuthCookie(c)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func getCurrentUserHandler(db *gorm.DB) gin.HandlerFunc {
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

		var user User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			return
		}

		c.JSON(http.StatusOK, sanitizeUser(&user))
	}
}

func checkFirstBootHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64
		db.Model(&User{}).Count(&count)

		c.JSON(http.StatusOK, gin.H{"needsSetup": count == 0})
	}
}

func setupSuperUserHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var count int64
		db.Model(&User{}).Count(&count)
		if count > 0 {
			c.JSON(http.StatusForbidden, gin.H{"message": "Super user already exists"})
			return
		}

		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}

		if req.Name == "" || req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Name, email and password are required"})
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
			Role:         RoleSuperAdmin,
			IsActive:     true,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
			return
		}

		token, err := generateToken(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
			return
		}

		setAuthCookie(c, token)

		c.JSON(http.StatusCreated, gin.H{
			"user":  sanitizeUser(&user),
			"token": token,
		})
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
