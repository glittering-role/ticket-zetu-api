package mail

import (
	"os"
	"strconv"

	"ticket-zetu-api/logs/handler"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

// EmailTemplateConfig holds email template paths
type EmailTemplateConfig struct {
	VerificationTemplatePath  string
	LoginWarningTemplatePath  string
	PasswordResetTemplatePath string
}

// AppConfig holds application URLs
type AppConfig struct {
	SecurityURL string
	SupportURL  string
	PrivacyURL  string
	TermsURL    string
}

// Config holds all email-related configurations
type Config struct {
	SMTPConfig     EmailConfig
	TemplateConfig EmailTemplateConfig
	AppConfig      AppConfig
	logHandler     *handler.LogHandler
}

// NewConfig initializes email configuration from environment variables
func NewConfig(logHandler *handler.LogHandler) (*Config, error) {
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		logHandler.LogError(nil, err, 0)
		return nil, err
	}

	config := &Config{
		SMTPConfig: EmailConfig{
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     port,
			SMTPUsername: os.Getenv("SMTP_USERNAME"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    os.Getenv("FROM_EMAIL"),
		},
		TemplateConfig: EmailTemplateConfig{
			VerificationTemplatePath:  os.Getenv("VERIFICATION_TEMPLATE_PATH"),
			LoginWarningTemplatePath:  os.Getenv("LOGIN_WARNING_TEMPLATE_PATH"),
			PasswordResetTemplatePath: os.Getenv("PASSWORD_RESET_TEMPLATE_PATH"),
		},
		AppConfig: AppConfig{
			SecurityURL: os.Getenv("SECURITY_URL"),
			SupportURL:  os.Getenv("SUPPORT_URL"),
			PrivacyURL:  os.Getenv("PRIVACY_URL"),
			TermsURL:    os.Getenv("TERMS_URL"),
		},
		logHandler: logHandler,
	}

	// Validate SMTP and template configurations
	if config.SMTPConfig.SMTPHost == "" ||
		config.SMTPConfig.SMTPUsername == "" ||
		config.SMTPConfig.SMTPPassword == "" ||
		config.SMTPConfig.FromEmail == "" ||
		config.TemplateConfig.VerificationTemplatePath == "" ||
		config.TemplateConfig.LoginWarningTemplatePath == "" ||
		config.TemplateConfig.PasswordResetTemplatePath == "" {
		logHandler.LogError(nil, err, 0)
		return nil, err
	}

	// Validate app URLs
	if config.AppConfig.SecurityURL == "" ||
		config.AppConfig.SupportURL == "" ||
		config.AppConfig.PrivacyURL == "" ||
		config.AppConfig.TermsURL == "" {
		logHandler.LogError(nil, err, 0)
		return nil, err
	}

	return config, nil
}

// GetSMTPConfig returns the SMTP configuration
func (c *Config) GetSMTPConfig() EmailConfig {
	return c.SMTPConfig
}

// GetTemplateConfig returns the template configuration
func (c *Config) GetTemplateConfig() EmailTemplateConfig {
	return c.TemplateConfig
}

// GetAppConfig returns the application URLs configuration
func (c *Config) GetAppConfig() AppConfig {
	return c.AppConfig
}
