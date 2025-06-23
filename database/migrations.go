package database

import (
	"log"
	Logs "ticket-zetu-api/logs/model"
	Category "ticket-zetu-api/modules/events/models/categories"
	Subcategory "ticket-zetu-api/modules/events/models/categories"

	OrganizationSubscription "ticket-zetu-api/modules/organizers/models"
	Organizer "ticket-zetu-api/modules/organizers/models"

	Permission "ticket-zetu-api/modules/users/models/authorization"
	Role "ticket-zetu-api/modules/users/models/authorization"
	RolePermission "ticket-zetu-api/modules/users/models/authorization"
	User "ticket-zetu-api/modules/users/models/members"
	UserLocation "ticket-zetu-api/modules/users/models/members"
	UserPreferences "ticket-zetu-api/modules/users/models/members"
	UserSecurityAttributes "ticket-zetu-api/modules/users/models/members"
	UserSession "ticket-zetu-api/modules/users/models/members"

	Comment "ticket-zetu-api/modules/events/models/events"
	Event "ticket-zetu-api/modules/events/models/events"
	EventImage "ticket-zetu-api/modules/events/models/events"
	Favorite "ticket-zetu-api/modules/events/models/events"
	Venue "ticket-zetu-api/modules/events/models/events"
	Vote "ticket-zetu-api/modules/events/models/events"
	Seat "ticket-zetu-api/modules/events/models/seats"
	SeatReservation "ticket-zetu-api/modules/events/models/seats"

	DiscountCode "ticket-zetu-api/modules/tickets/models/tickets"
	PriceTier "ticket-zetu-api/modules/tickets/models/tickets"
	Ticket "ticket-zetu-api/modules/tickets/models/tickets"
	TicketType "ticket-zetu-api/modules/tickets/models/tickets"

	VenueImage "ticket-zetu-api/modules/events/models/events"

	Notification "ticket-zetu-api/modules/notifications/models"
	UserNotification "ticket-zetu-api/modules/notifications/models"

	ArtistProfile "ticket-zetu-api/modules/users/models/artist"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	// List all your models here in logical groups
	models := []interface{}{
		// System Models
		&Logs.Log{},

		// Authorization Models
		&Permission.Permission{},
		&Role.Role{},
		&RolePermission.RolePermission{},

		// User Models
		&User.User{},
		&UserSecurityAttributes.UserSecurityAttributes{},
		&UserSession.UserSession{},
		&UserPreferences.UserPreferences{},
		&UserLocation.UserLocation{},

		// Category Models
		&Category.Category{},
		&Subcategory.Subcategory{},

		// Organizer Models
		&Organizer.Organizer{},
		&OrganizationSubscription.OrganizationSubscription{},

		// Event Models
		&Venue.Venue{},
		&VenueImage.VenueImage{},
		&Event.Event{},
		&EventImage.EventImage{},
		&Favorite.Favorite{},
		&Vote.Vote{},
		&Comment.Comment{},
		&Seat.Seat{},
		&SeatReservation.SeatReservation{},

		// Ticket Models
		&PriceTier.PriceTier{},
		&TicketType.TicketType{},
		&DiscountCode.DiscountCode{},
		&Ticket.Ticket{},

		//Notification
		&Notification.Notification{},
		&UserNotification.UserNotification{},

		//ArtistProfile
		&ArtistProfile.ArtistProfile{},
	}

	db = db.Debug()

	// Check if any migrations are needed
	migrationsNeeded := false
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			migrationsNeeded = true
			break
		}
	}

	if !migrationsNeeded {
		log.Println("No migrations needed - all tables exist")
		return nil
	}

	log.Println("Running database migrations...")
	migrationCount := 0
	for _, model := range models {
		if !db.Migrator().HasTable(model) {
			if err := db.AutoMigrate(model); err != nil {
				return err
			}
			migrationCount++
			log.Printf("Migrated table for %T\n", model)
		}
	}

	if migrationCount == 0 {
		log.Println("No new migrations were applied")
	} else {
		log.Printf("Migrations completed successfully. Applied %d migrations\n", migrationCount)
	}
	return nil
}
