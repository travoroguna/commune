package main

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleSuperAdmin      UserRole = "super_admin"
	RoleAdmin           UserRole = "admin"
	RoleModerator       UserRole = "moderator"
	RoleServiceProvider UserRole = "service_provider"
	RoleUser            UserRole = "user"
)

// User represents a user in the system with authentication and role information
type User struct {
	gorm.Model
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         UserRole `gorm:"type:varchar(50);default:'user';not null"`
	IsActive     bool     `gorm:"default:true;not null"`

	// Relationships
	Communities     []Community      `gorm:"many2many:user_communities;"`
	Posts           []Post           `gorm:"foreignKey:AuthorID"`
	ServiceRequests []ServiceRequest `gorm:"foreignKey:RequesterID"`
	ServiceOffers   []ServiceOffer   `gorm:"foreignKey:ProviderID"`
	Comments        []Comment        `gorm:"foreignKey:AuthorID"`
	Ratings         []Rating         `gorm:"foreignKey:RaterID"`
	ReceivedRatings []Rating         `gorm:"foreignKey:ProviderID"`
}

// Community represents a community (e.g., apartment complex, estate)
// Each community lives in its own space and can have its own domain
type Community struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Slug        string `gorm:"uniqueIndex;not null"` // URL-friendly identifier (e.g., "sunset-apartments")
	Description string `gorm:"type:text"`

	// Domain configuration for multi-tenancy
	Subdomain   string `gorm:"uniqueIndex"` // Subdomain for community (e.g., "sunset" -> sunset.commune.com)
	CustomDomain string `gorm:"uniqueIndex"` // Custom domain (e.g., "sunset-apts.com")

	// Location information
	Address     string
	City        string
	State       string
	Country     string
	ZipCode     string

	IsActive    bool   `gorm:"default:true;not null"`

	// Relationships
	Users           []User           `gorm:"many2many:user_communities;"`
	Posts           []Post           `gorm:"foreignKey:CommunityID"`
	ServiceRequests []ServiceRequest `gorm:"foreignKey:CommunityID"`
}

// UserCommunity represents the many-to-many relationship between users and communities
// with additional metadata about the user's role in that specific community
type UserCommunity struct {
	UserID      uint      `gorm:"primaryKey"`
	CommunityID uint      `gorm:"primaryKey"`
	Role        UserRole  `gorm:"type:varchar(50);default:'user';not null"`
	JoinedAt    time.Time `gorm:"autoCreateTime"`
	IsActive    bool      `gorm:"default:true;not null"`

	// Foreign keys
	User      User      `gorm:"foreignKey:UserID"`
	Community Community `gorm:"foreignKey:CommunityID"`
}

// Post represents content posted by users in a community
type Post struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Content     string `gorm:"type:text;not null"`
	AuthorID    uint   `gorm:"not null;index"`
	CommunityID uint   `gorm:"not null;index"`
	IsPublished bool   `gorm:"default:true;not null"`
	ViewCount   int    `gorm:"default:0"`

	// Relationships
	Author    User      `gorm:"foreignKey:AuthorID"`
	Community Community `gorm:"foreignKey:CommunityID"`
	Comments  []Comment `gorm:"foreignKey:PostID"`
}

// ServiceRequest represents a request for a service in a community
// Note: AcceptedOfferID creates a bidirectional relationship with ServiceOffer.
// When accepting an offer, update both ServiceRequest.AcceptedOfferID and ServiceRequest.Status
// When deleting a ServiceRequest, associated ServiceOffers will need to be handled (cascade or set null)
type ServiceRequest struct {
	gorm.Model
	Title          string `gorm:"not null"`
	Description    string `gorm:"type:text;not null"`
	Category       string `gorm:"index"`
	RequesterID    uint   `gorm:"not null;index"`
	CommunityID    uint   `gorm:"not null;index"`
	Status         string `gorm:"type:varchar(50);default:'open';not null;index"` // open, in_progress, completed, cancelled
	Budget         float64
	AcceptedOfferID *uint  `gorm:"index"` // References ServiceOffer.ID - nullable until offer is accepted
	CompletedAt    *time.Time

	// Relationships
	Requester      User           `gorm:"foreignKey:RequesterID"`
	Community      Community      `gorm:"foreignKey:CommunityID"`
	ServiceOffers  []ServiceOffer `gorm:"foreignKey:ServiceRequestID"`
	Comments       []Comment      `gorm:"foreignKey:ServiceRequestID"`
	AcceptedOffer  *ServiceOffer  `gorm:"foreignKey:AcceptedOfferID;constraint:OnDelete:SET NULL"` // Set to NULL if offer is deleted
}

// ServiceOffer represents an offer by a service provider for a service request
type ServiceOffer struct {
	gorm.Model
	ServiceRequestID uint   `gorm:"not null;index"`
	ProviderID       uint   `gorm:"not null;index"`
	Description      string `gorm:"type:text;not null"`
	ProposedPrice    float64
	EstimatedDuration string
	Status           string `gorm:"type:varchar(50);default:'pending';not null"` // pending, accepted, rejected, withdrawn

	// Relationships
	ServiceRequest ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
	Provider       User           `gorm:"foreignKey:ProviderID"`
	Comments       []Comment      `gorm:"foreignKey:ServiceOfferID"`
}

// Comment represents a comment on a post, service request, or service offer
type Comment struct {
	gorm.Model
	Content          string `gorm:"type:text;not null"`
	AuthorID         uint   `gorm:"not null;index"`
	PostID           *uint  `gorm:"index"`
	ServiceRequestID *uint  `gorm:"index"`
	ServiceOfferID   *uint  `gorm:"index"`
	ParentCommentID  *uint  `gorm:"index"` // For nested comments/replies

	// Relationships
	Author         User            `gorm:"foreignKey:AuthorID"`
	Post           *Post           `gorm:"foreignKey:PostID"`
	ServiceRequest *ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
	ServiceOffer   *ServiceOffer   `gorm:"foreignKey:ServiceOfferID"`
	ParentComment  *Comment        `gorm:"foreignKey:ParentCommentID"`
	Replies        []Comment       `gorm:"foreignKey:ParentCommentID"`
}

// Rating represents a rating and review for a service provider
type Rating struct {
	gorm.Model
	ProviderID       uint   `gorm:"not null;index"`
	RaterID          uint   `gorm:"not null;index"`
	ServiceRequestID uint   `gorm:"not null;index"`
	Score            int    `gorm:"not null;check:score >= 1 AND score <= 5"` // 1-5 stars with database constraint
	Review           string `gorm:"type:text"`

	// Relationships
	Provider       User           `gorm:"foreignKey:ProviderID"`
	Rater          User           `gorm:"foreignKey:RaterID"`
	ServiceRequest ServiceRequest `gorm:"foreignKey:ServiceRequestID"`
}

// JoinRequest represents a request to join a community
type JoinRequest struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index"`
	CommunityID uint   `gorm:"not null;index"`
	Status      string `gorm:"type:varchar(50);default:'pending';not null;index"` // pending, approved, rejected
	Message     string `gorm:"type:text"`

	// Relationships
	User      User      `gorm:"foreignKey:UserID"`
	Community Community `gorm:"foreignKey:CommunityID"`
}
