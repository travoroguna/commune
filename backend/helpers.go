package main

import (
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// GenerateSlug creates a URL-friendly slug from a string
// Example: "Sunset Apartments" -> "sunset-apartments"
func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	
	// Remove leading and trailing hyphens
	s = strings.Trim(s, "-")
	
	return s
}

// GetCommunityByDomain finds a community by custom domain or subdomain
// This will be used to route requests to the correct community
func GetCommunityByDomain(db *gorm.DB, domain string) (*Community, error) {
	var community Community
	
	// Check if it's a custom domain
	err := db.Where("custom_domain = ? AND is_active = ?", domain, true).First(&community).Error
	if err == nil {
		return &community, nil
	}
	
	// Check if it's a subdomain (extract subdomain part)
	// Example: "sunset.commune.com" -> "sunset"
	parts := strings.Split(domain, ".")
	if len(parts) > 0 {
		subdomain := parts[0]
		err = db.Where("subdomain = ? AND is_active = ?", subdomain, true).First(&community).Error
		if err == nil {
			return &community, nil
		}
	}
	
	return nil, err
}

// GetCommunityBySlug finds a community by its slug
func GetCommunityBySlug(db *gorm.DB, slug string) (*Community, error) {
	var community Community
	err := db.Where("slug = ? AND is_active = ?", slug, true).First(&community).Error
	return &community, err
}
