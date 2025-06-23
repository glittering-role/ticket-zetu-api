package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"ticket-zetu-api/config"
)

var DB *gorm.DB

// InitDB initializes the database connection and sets up the connection pool
func InitDB() {
	// Load the configuration
	appConfig := config.LoadConfig()
	dsn := config.GetDSN(appConfig)

	// Open the database connection using gorm.io/gorm and the mysql driver
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	// Setting up connection pool (gorm.io/gorm uses *sql.DB for connection pooling)
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error getting the sql.DB instance from gorm.DB: %v", err)
	}

	// Setting the connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * 60 * 1000)

	fmt.Println("Database connected successfully with connection pooling enabled.")
}

// CloseDB closes the database connection
func CloseDB() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error getting the sql.DB instance from gorm.DB: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Fatalf("Error closing the database connection: %v", err)
	}
}
