package main

// This file contains example usage patterns for the database models.
// These examples can be used as reference when implementing API endpoints.

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Example 1: Create a new community (Super Admin only)
func ExampleCreateCommunity(db *gorm.DB, adminID uint) (*Community, error) {
	name := "Sunrise Apartments"
	community := Community{
		Name:        name,
		Slug:        GenerateSlug(name), // "sunrise-apartments"
		Subdomain:   "sunrise",          // Will be accessible at sunrise.commune.com
		Description: "A modern apartment complex with 200+ units",
		Address:     "456 Oak Avenue",
		City:        "Los Angeles",
		State:       "CA",
		Country:     "USA",
		ZipCode:     "90001",
		IsActive:    true,
	}

	result := db.Create(&community)
	return &community, result.Error
}

// Example 2: User joins a community
func ExampleJoinCommunity(db *gorm.DB, userID, communityID uint) error {
	userCommunity := UserCommunity{
		UserID:      userID,
		CommunityID: communityID,
		Role:        RoleUser,
		JoinedAt:    time.Now(),
		IsActive:    true,
	}

	return db.Create(&userCommunity).Error
}

// Example 3: Create a post in a community
func ExampleCreatePost(db *gorm.DB, authorID, communityID uint, title, content string) (*Post, error) {
	post := Post{
		Title:       title,
		Content:     content,
		AuthorID:    authorID,
		CommunityID: communityID,
		IsPublished: true,
		ViewCount:   0,
	}

	result := db.Create(&post)
	return &post, result.Error
}

// Example 4: Create a service request
func ExampleCreateServiceRequest(db *gorm.DB, requesterID, communityID uint) (*ServiceRequest, error) {
	request := ServiceRequest{
		Title:       "Need electrician for outlet repair",
		Description: "Several outlets in my apartment are not working. Need urgent repair.",
		Category:    "Electrical",
		RequesterID: requesterID,
		CommunityID: communityID,
		Status:      "open",
		Budget:      200.00,
	}

	result := db.Create(&request)
	return &request, result.Error
}

// Example 5: Service provider creates an offer
func ExampleCreateServiceOffer(db *gorm.DB, providerID, serviceRequestID uint) (*ServiceOffer, error) {
	offer := ServiceOffer{
		ServiceRequestID:  serviceRequestID,
		ProviderID:        providerID,
		Description:       "I'm a licensed electrician with 15 years of experience. I can fix this today.",
		ProposedPrice:     150.00,
		EstimatedDuration: "2-3 hours",
		Status:            "pending",
	}

	result := db.Create(&offer)
	return &offer, result.Error
}

// Example 6: Accept a service offer
func ExampleAcceptServiceOffer(db *gorm.DB, requestID, offerID uint) error {
	// Update service request
	updates := map[string]interface{}{
		"accepted_offer_id": offerID,
		"status":           "in_progress",
	}

	if err := db.Model(&ServiceRequest{}).Where("id = ?", requestID).Updates(updates).Error; err != nil {
		return err
	}

	// Update offer status
	return db.Model(&ServiceOffer{}).Where("id = ?", offerID).Update("status", "accepted").Error
}

// Example 7: Complete service and add rating
func ExampleCompleteServiceAndRate(db *gorm.DB, requestID, providerID, raterID uint, score int, review string) error {
	// Start transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// Mark service as completed
		completedAt := time.Now()
		if err := tx.Model(&ServiceRequest{}).Where("id = ?", requestID).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": completedAt,
		}).Error; err != nil {
			return err
		}

		// Create rating
		rating := Rating{
			ProviderID:       providerID,
			RaterID:          raterID,
			ServiceRequestID: requestID,
			Score:            score,
			Review:           review,
		}

		return tx.Create(&rating).Error
	})
}

// Example 8: Add a comment to a post
func ExampleAddCommentToPost(db *gorm.DB, authorID, postID uint, content string) (*Comment, error) {
	comment := Comment{
		Content:  content,
		AuthorID: authorID,
		PostID:   &postID,
	}

	result := db.Create(&comment)
	return &comment, result.Error
}

// Example 9: Reply to a comment (nested comment)
func ExampleReplyToComment(db *gorm.DB, authorID, parentCommentID uint, content string) (*Comment, error) {
	comment := Comment{
		Content:         content,
		AuthorID:        authorID,
		ParentCommentID: &parentCommentID,
	}

	result := db.Create(&comment)
	return &comment, result.Error
}

// Example 10: Get all posts in a community with author information
func ExampleGetCommunityPosts(db *gorm.DB, communityID uint) ([]Post, error) {
	var posts []Post
	err := db.Preload("Author").
		Where("community_id = ? AND is_published = ?", communityID, true).
		Order("created_at DESC").
		Find(&posts).Error

	return posts, err
}

// Example 11: Get service requests with offers
func ExampleGetServiceRequestsWithOffers(db *gorm.DB, communityID uint, status string) ([]ServiceRequest, error) {
	var requests []ServiceRequest
	query := db.Preload("ServiceOffers").
		Preload("ServiceOffers.Provider").
		Preload("Requester").
		Where("community_id = ?", communityID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// Example 12: Get service provider's average rating
func ExampleGetProviderAverageRating(db *gorm.DB, providerID uint) (float64, int, error) {
	var result struct {
		AvgScore float64
		Count    int
	}

	err := db.Model(&Rating{}).
		Select("AVG(score) as avg_score, COUNT(*) as count").
		Where("provider_id = ?", providerID).
		Scan(&result).Error

	return result.AvgScore, result.Count, err
}

// Example 13: Get user's communities with their roles
func ExampleGetUserCommunities(db *gorm.DB, userID uint) ([]UserCommunity, error) {
	var userCommunities []UserCommunity
	err := db.Preload("Community").
		Where("user_id = ? AND is_active = ?", userID, true).
		Find(&userCommunities).Error

	return userCommunities, err
}

// Example 14: Search service requests by category
func ExampleSearchServiceRequestsByCategory(db *gorm.DB, communityID uint, category string) ([]ServiceRequest, error) {
	var requests []ServiceRequest
	err := db.Preload("Requester").
		Where("community_id = ? AND category = ? AND status = ?", communityID, category, "open").
		Order("created_at DESC").
		Find(&requests).Error

	return requests, err
}

// Example 15: Get post with all comments (including nested)
func ExampleGetPostWithComments(db *gorm.DB, postID uint) (*Post, error) {
	var post Post
	err := db.Preload("Comments").
		Preload("Comments.Author").
		Preload("Comments.Replies").
		Preload("Comments.Replies.Author").
		Where("id = ?", postID).
		First(&post).Error

	return &post, err
}

// Example 16: Promote user to moderator in a community
func ExamplePromoteUserToModerator(db *gorm.DB, userID, communityID uint) error {
	return db.Model(&UserCommunity{}).
		Where("user_id = ? AND community_id = ?", userID, communityID).
		Update("role", RoleModerator).Error
}

// Example 17: Get all service providers in a community with ratings
func ExampleGetServiceProvidersWithRatings(db *gorm.DB, communityID uint) ([]User, error) {
	var providers []User

	err := db.Joins("JOIN user_communities ON user_communities.user_id = users.id").
		Preload("ReceivedRatings", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(5)
		}).
		Where("user_communities.community_id = ? AND user_communities.role = ? AND user_communities.is_active = ?",
			communityID, RoleServiceProvider, true).
		Find(&providers).Error

	return providers, err
}

// Example 18: Get recent activity in a community
func ExampleGetCommunityActivity(db *gorm.DB, communityID uint, limit int) (interface{}, error) {
	type Activity struct {
		Type      string
		ID        uint
		Title     string
		UserName  string
		CreatedAt time.Time
	}

	var activities []Activity

	// Get recent posts
	var posts []Post
	db.Preload("Author").
		Where("community_id = ?", communityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&posts)

	for _, post := range posts {
		activities = append(activities, Activity{
			Type:      "post",
			ID:        post.ID,
			Title:     post.Title,
			UserName:  post.Author.Name,
			CreatedAt: post.CreatedAt,
		})
	}

	// Get recent service requests
	var requests []ServiceRequest
	db.Preload("Requester").
		Where("community_id = ?", communityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&requests)

	for _, req := range requests {
		activities = append(activities, Activity{
			Type:      "service_request",
			ID:        req.ID,
			Title:     req.Title,
			UserName:  req.Requester.Name,
			CreatedAt: req.CreatedAt,
		})
	}

	return activities, nil
}

// Example 19: Cancel a service request
func ExampleCancelServiceRequest(db *gorm.DB, requestID, userID uint) error {
	// Verify the user is the requester
	var request ServiceRequest
	if err := db.First(&request, requestID).Error; err != nil {
		return err
	}

	if request.RequesterID != userID {
		return fmt.Errorf("unauthorized: only requester can cancel")
	}

	return db.Model(&ServiceRequest{}).
		Where("id = ?", requestID).
		Update("status", "cancelled").Error
}

// Example 20: Soft delete a post
func ExampleDeletePost(db *gorm.DB, postID, authorID uint) error {
	// Verify the user is the author
	var post Post
	if err := db.First(&post, postID).Error; err != nil {
		return err
	}

	if post.AuthorID != authorID {
		return fmt.Errorf("unauthorized: only author can delete")
	}

	// GORM's Delete performs soft delete
	return db.Delete(&Post{}, postID).Error
}

// Example 21: Create community with custom domain
func ExampleCreateCommunityWithCustomDomain(db *gorm.DB) (*Community, error) {
	name := "Luxury Heights"
	community := Community{
		Name:         name,
		Slug:         GenerateSlug(name),      // "luxury-heights"
		Subdomain:    "luxuryheights",         // luxuryheights.commune.com
		CustomDomain: "luxuryheights.com",     // Custom domain
		Description:  "Premium luxury apartments",
		City:         "Miami",
		State:        "FL",
		Country:      "USA",
		IsActive:     true,
	}

	result := db.Create(&community)
	return &community, result.Error
}

// Example 22: Route request based on domain
func ExampleRouteByDomain(db *gorm.DB, requestDomain string) (*Community, error) {
	// This would be called in HTTP middleware to determine which community
	// the request is for based on the domain
	return GetCommunityByDomain(db, requestDomain)
}

// Example 23: Route request based on slug (for shared domain)
func ExampleRouteBySlug(db *gorm.DB, slug string) (*Community, error) {
	// This would be used for URL paths like: commune.com/c/sunset-apartments
	return GetCommunityBySlug(db, slug)
}
