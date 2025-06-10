package mail_service

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	"ticket-zetu-api/mail"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

// SendLoginWarning queues a login warning email
func (s *emailService) SendLoginWarning(c *fiber.Ctx, email, username, deviceInfo, ipAddress string, loginTime time.Time, warningType string) error {
	smtpConfig := s.config.GetSMTPConfig()
	templateConfig := s.config.GetTemplateConfig()
	appConfig := s.config.GetAppConfig()

	job := emailJob{
		ctx: c,
		execute: func() error {
			return s.sendLoginWarningEmail(
				email,
				username,
				deviceInfo,
				ipAddress,
				loginTime,
				warningType,
				smtpConfig,
				templateConfig,
				appConfig,
			)
		},
	}

	select {
	case s.jobQueue <- job:
		log.Printf("Enqueued login warning email job for %s", email)
		return nil
	case <-time.After(100 * time.Millisecond):
		log.Printf("Email queue overloaded for %s", email)
		return fmt.Errorf("email queue overloaded")
	}
}

func (s *emailService) sendLoginWarningEmail(
	email, username, deviceInfo, ipAddress string,
	loginTime time.Time,
	warningType string,
	smtpConfig mail.EmailConfig,
	templateConfig mail.EmailTemplateConfig,
	appConfig mail.AppConfig,
) error {
	var subject, warningMessage string
	switch warningType {
	case "new_login":
		subject = "New Login Detected"
		warningMessage = "We detected a new login to your account."
	case "lockout_failed_attempts":
		subject = "Account Locked: Too Many Failed Attempts"
		warningMessage = "Your account has been locked due to too many failed login attempts."
	default:
		subject = "Security Alert"
		warningMessage = "A security event was detected on your account."
	}

	ticketNumber := fmt.Sprintf("ALRT-%d-%d", time.Now().Unix()%10000, randInt(100, 999))

	data := struct {
		Username       string
		TicketNumber   string
		DeviceInfo     string
		IPAddress      string
		LoginTime      string
		WarningMessage string
		SecurityURL    string
		SupportURL     string
		PrivacyURL     string
		TermsURL       string
	}{
		Username:       username,
		TicketNumber:   ticketNumber,
		DeviceInfo:     deviceInfo,
		IPAddress:      ipAddress,
		LoginTime:      loginTime.Format("2006-01-02 15:04:05"),
		WarningMessage: warningMessage,
		SecurityURL:    appConfig.SecurityURL,
		SupportURL:     appConfig.SupportURL,
		PrivacyURL:     appConfig.PrivacyURL,
		TermsURL:       appConfig.TermsURL,
	}

	log.Printf("Loading login warning template from: %s", templateConfig.LoginWarningTemplatePath)
	templateContent, err := os.ReadFile(templateConfig.LoginWarningTemplatePath)
	if err != nil {
		log.Printf("Failed to read login warning template: %v", err)
		return fmt.Errorf("failed to read template: %w", err)
	}
	log.Printf("Login warning template content length: %d bytes", len(templateContent))

	var buf bytes.Buffer
	tmpl, err := template.New("loginWarningEmail").Parse(string(templateContent))
	if err != nil {
		log.Printf("Login warning template parsing failed: %v", err)
		return fmt.Errorf("template parsing failed: %w", err)
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Login warning template execution failed: %v", err)
		return fmt.Errorf("template execution failed: %w", err)
	}
	log.Printf("Rendered login warning email length: %d bytes", buf.Len())

	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", buf.String())

	log.Printf("Sending login warning email to %s from %s via %s:%d", email, smtpConfig.FromEmail, smtpConfig.SMTPHost, smtpConfig.SMTPPort)
	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send login warning email to %s: %v", email, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("Login warning email sent successfully to %s", email)
	return nil
}
