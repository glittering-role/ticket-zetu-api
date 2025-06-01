package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"ticket-zetu-api/cloudinary"
)

type AppConfig struct {
	Port       string
	Env        string
	AppName    string
	DBName     string
	DBUser     string
	DBPass     string
	DBHost     string
	Dialect    string
	Cloudinary cloudinary.Config
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *AppConfig {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Create an AppConfig instance and populate it with values from the environment
	return &AppConfig{
		Port:    getEnv("PORT", "8080"),
		Env:     getEnv("GO_ENV", "development"),
		AppName: getEnv("APP_NAME", "ticket-zetu-api"),
		DBName:  getEnv("DB_NAME", ""),
		DBUser:  getEnv("DB_USER", ""),
		DBPass:  getEnv("DB_PASSWORD", ""),
		DBHost:  getEnv("DB_HOST", "localhost:3306"),
		Dialect: getEnv("DB_DIALECT", "mysql"),
		Cloudinary: cloudinary.Config{
			CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:    getEnv("CLOUDINARY_API_KEY", ""),
			APISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		},
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

// getEnv retrieves an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
