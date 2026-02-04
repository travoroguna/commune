package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Connect to database
	db, err := gorm.Open(sqlite.Open("commune.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create community
	community := Community{
		Name:        "Sunset Apartments",
		Slug:        "sunset-apartments",
		Description: "A beautiful apartment complex with amenities",
		City:        "Los Angeles",
		State:       "CA",
		Country:     "USA",
		IsActive:    true,
	}
	if err := db.Create(&community).Error; err != nil {
		log.Fatal("Failed to create community:", err)
	}
	fmt.Printf("Created community: %s (ID: %d)\n", community.Name, community.ID)

	// Create service requests
	services := []ServiceRequest{
		{
			Title:       "Need plumber for kitchen sink",
			Description: "Kitchen sink is leaking and needs immediate repair. Water is dripping constantly.",
			Category:    "Plumbing",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      150.00,
		},
		{
			Title:       "Electrician needed for outlet repair",
			Description: "Several outlets in the living room are not working. Need a licensed electrician to fix them.",
			Category:    "Electrical",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      200.00,
		},
		{
			Title:       "Carpet cleaning service",
			Description: "Need professional carpet cleaning for a 3-bedroom apartment. Prefer eco-friendly products.",
			Category:    "Cleaning",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      120.00,
		},
		{
			Title:       "AC maintenance required",
			Description: "Annual AC maintenance and filter replacement needed before summer.",
			Category:    "HVAC",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "in_progress",
			Budget:      180.00,
		},
		{
			Title:       "Painting service for bedroom",
			Description: "Looking for a professional painter to paint a master bedroom. Need color consultation as well.",
			Category:    "Painting",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      300.00,
		},
		{
			Title:       "Locksmith service",
			Description: "Need to replace locks on main entrance door. Previous tenant left with keys.",
			Category:    "Security",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "completed",
			Budget:      100.00,
		},
		{
			Title:       "Appliance repair - Refrigerator",
			Description: "Refrigerator is making loud noises and not cooling properly. Needs diagnosis and repair.",
			Category:    "Appliance Repair",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      250.00,
		},
		{
			Title:       "Pest control service",
			Description: "Noticed some ants in the kitchen. Need pest control service preferably with pet-safe products.",
			Category:    "Pest Control",
			RequesterID: 1,
			CommunityID: community.ID,
			Status:      "open",
			Budget:      90.00,
		},
	}

	for _, service := range services {
		if err := db.Create(&service).Error; err != nil {
			log.Printf("Failed to create service: %v", err)
			continue
		}
		fmt.Printf("Created service: %s (ID: %d, Status: %s)\n", service.Title, service.ID, service.Status)
	}

	fmt.Println("\nTest data created successfully!")
	os.Exit(0)
}
