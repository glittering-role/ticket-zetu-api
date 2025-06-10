package main

import (
	"errors"
	"log"
	"time"

	"ticket-zetu-api/config"
	"ticket-zetu-api/database"
	model "ticket-zetu-api/modules/users/models/authorization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) error {
	roles := []model.Role{
		{
			ID:             uuid.New().String(),
			RoleName:       "superadmin",
			Description:    "Super Administrator with full system access",
			Level:          100,
			Status:         model.RoleActive,
			IsSystemRole:   true,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "admin",
			Description:    "Administrator with full access",
			Level:          50,
			Status:         model.RoleActive,
			IsSystemRole:   true,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "artist",
			Description:    "Artist with special permissions for content creation",
			Level:          40,
			Status:         model.RoleActive,
			IsSystemRole:   false,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "ticket_curator",
			Description:    "Can favorite and organize tickets in special collections",
			Level:          30,
			Status:         model.RoleActive,
			IsSystemRole:   false,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "ticket_follower",
			Description:    "Can follow and favorite tickets",
			Level:          20,
			Status:         model.RoleActive,
			IsSystemRole:   false,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "user",
			Description:    "Standard user with basic access",
			Level:          10,
			Status:         model.RoleActive,
			IsSystemRole:   false,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
		{
			ID:             uuid.New().String(),
			RoleName:       "guest",
			Description:    "Guest user with limited access",
			Level:          1,
			Status:         model.RoleActive,
			IsSystemRole:   false,
			NumberOfUsers:  0,
			CreatedBy:      "system",
			LastModifiedBy: "system",
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Version:        1,
		},
	}

	for _, role := range roles {
		var existingRole model.Role
		if err := db.Where("role_name = ?", role.RoleName).First(&existingRole).Error; err == nil {
			log.Printf("Role '%s' already exists, skipping...", role.RoleName)
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err := db.Create(&role).Error; err != nil {
			return err
		}
		log.Printf("Created role '%s' with ID: %s", role.RoleName, role.ID)
	}

	return nil
}

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	database.InitDB()
	defer database.CloseDB()

	// Run the seed
	log.Println("Starting role seeding...")
	if err := SeedRoles(database.DB); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}

	log.Println("Role seeding completed successfully!")
}
