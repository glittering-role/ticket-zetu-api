package database

import (
	"gorm.io/gorm"
	"log"
	Logs "ticket-zetu-api/logs/model"
	Role "ticket-zetu-api/modules/users/models/authorization"
)

func Migrate(db *gorm.DB) error {
	// List all your models here
	models := []interface{}{
		&Logs.Log{},

		//Users Module
		&Role.Role{},
	}

	// Enable detailed logging for migrations
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
