package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// AppConfig holds the application configuration
type AppConfig struct {
	Port    string
	Env     string
	AppName string
	DBName  string
	DBUser  string
	DBPass  string
	DBHost  string
	Dialect string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *AppConfig {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create an AppConfig instance and populate it with values from the environment
	return &AppConfig{
		Port:    os.Getenv("PORT"),
		Env:     os.Getenv("GO_ENV"),
		AppName: os.Getenv("APP_NAME"),
		DBName:  os.Getenv("DB_NAME"),
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASSWORD"),
		DBHost:  os.Getenv("DB_HOST"),
		Dialect: os.Getenv("DB_DIALECT"),
	}
}

// GetDSN generates the MySQL Data Source Name (DSN) string
func GetDSN(config *AppConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.DBUser,
		config.DBPass,
		config.DBHost,
		config.DBName,
	)
}
