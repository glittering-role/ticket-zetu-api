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

func (s *emailService) SendPasswordResetEmail(c *fiber.Ctx, email, username, resetToken string) error {
	smtpConfig := s.config.GetSMTPConfig()
	templateConfig := s.config.GetTemplateConfig()
	appConfig := s.config.GetAppConfig()

	job := emailJob{
		ctx: c,
		execute: func() error {
			return s.sendPasswordResetEmail(
				email,
				username,
				resetToken,
				smtpConfig,
				templateConfig,
				appConfig,
			)
		},
	}

	select {
	case s.jobQueue <- job:
		log.Printf("Enqueued password reset email job for %s", email)
		return nil
	case <-time.After(100 * time.Millisecond):
		log.Printf("Email queue overloaded for %s", email)
		return fmt.Errorf("email queue overloaded")
	}
}

func (s *emailService) sendPasswordResetEmail(
	email, username, resetToken string,
	smtpConfig mail.EmailConfig,
	templateConfig mail.EmailTemplateConfig,
	appConfig mail.AppConfig,
) error {
	ticketNumber := fmt.Sprintf("RST-%d-%d", time.Now().Unix()%10000, randInt(100, 999))
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", appConfig.SecurityURL, resetToken)

	data := struct {
		Username     string
		TicketNumber string
		ResetURL     string
		SupportURL   string
		PrivacyURL   string
		TermsURL     string
		ExpiryTime   string
	}{
		Username:     username,
		TicketNumber: ticketNumber,
		ResetURL:     resetURL,
		SupportURL:   appConfig.SupportURL,
		PrivacyURL:   appConfig.PrivacyURL,
		TermsURL:     appConfig.TermsURL,
		ExpiryTime:   time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	log.Printf("Loading password reset template from: %s", templateConfig.PasswordResetTemplatePath)
	templateContent, err := os.ReadFile(templateConfig.PasswordResetTemplatePath)
	if err != nil {
		log.Printf("Failed to read password reset template: %v", err)
		return fmt.Errorf("failed to read template: %w", err)
	}
	log.Printf("Password reset template content length: %d bytes", len(templateContent))

	var buf bytes.Buffer
	tmpl, err := template.New("passwordResetEmail").Parse(string(templateContent))
	if err != nil {
		log.Printf("Password reset template parsing failed: %v", err)
		return fmt.Errorf("template parsing failed: %w", err)
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Password reset template execution failed: %v", err)
		return fmt.Errorf("template execution failed: %w", err)
	}
	log.Printf("Rendered password reset email length: %d bytes", buf.Len())

	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Reset Your Ticket Zetu Password")
	m.SetBody("text/html", buf.String())

	log.Printf("Sending password reset email to %s from %s via %s:%d", email, smtpConfig.FromEmail, smtpConfig.SMTPHost, smtpConfig.SMTPPort)
	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send password reset email to %s: %v", email, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("Password reset email sent successfully to %s", email)
	return nil
}
