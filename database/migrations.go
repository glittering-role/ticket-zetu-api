package database

import (
	"log"
	Logs "ticket-zetu-api/logs/model"
	Organizer "ticket-zetu-api/modules/organizers/models"
	Category "ticket-zetu-api/modules/tickets/models/categories"
	Subcategory "ticket-zetu-api/modules/tickets/models/categories"
	Permission "ticket-zetu-api/modules/users/models/authorization"
	Role "ticket-zetu-api/modules/users/models/authorization"
	RolePermission "ticket-zetu-api/modules/users/models/authorization"
	User "ticket-zetu-api/modules/users/models/members"
	UserLocation "ticket-zetu-api/modules/users/models/members"
	UserPreferences "ticket-zetu-api/modules/users/models/members"
	UserSecurityAttributes "ticket-zetu-api/modules/users/models/members"
	UserSession "ticket-zetu-api/modules/users/models/members"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// List all your models here
	models := []interface{}{
		&Logs.Log{},

		//Users Module
		&Role.Role{},
		&Permission.Permission{},
		&RolePermission.RolePermission{},
		&User.User{},
		&UserSecurityAttributes.UserSecurityAttributes{},
		&UserSession.UserSession{},
		&UserPreferences.UserPreferences{},
		&UserLocation.UserLocation{},
		&UserSecurityAttributes.UserSecurityAttributes{},

		//Category
		&Category.Category{},
		&Subcategory.Subcategory{},

		//Organizer
		&Organizer.Organizer{},
	}

	db = db.Debug()

	log.Println("Running database migrations...")
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}
	log.Println("Migrations completed successfully")
	return nil
}
