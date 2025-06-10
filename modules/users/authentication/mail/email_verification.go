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

// GenerateAndSendVerificationCode generates and queues an email verification code
func (s *emailService) GenerateAndSendVerificationCode(c *fiber.Ctx, email, username, userID string) (string, error) {
	code, err := s.generateVerificationCode()
	if err != nil {
		log.Printf("Verification code generation failed: %v", err)
		return "", fmt.Errorf("verification code generation failed: %w", err)
	}

	// Validate code length
	if len(code) != 8 {
		log.Printf("Invalid code length: %d for code %s", len(code), code)
		return "", fmt.Errorf("generated code is not 8 digits")
	}

	// Prepare all data needed for the email job
	smtpConfig := s.config.GetSMTPConfig()
	templateConfig := s.config.GetTemplateConfig()
	appConfig := s.config.GetAppConfig()

	job := emailJob{
		ctx: c,
		execute: func() error {
			return s.sendVerificationEmail(
				email,
				username,
				code,
				smtpConfig,
				templateConfig,
				appConfig,
			)
		},
	}

	// Non-blocking enqueue with timeout
	select {
	case s.jobQueue <- job:
		log.Printf("Enqueued email job for %s with code %s, queue length: %d", email, code, len(s.jobQueue))
		return code, nil
	case <-time.After(100 * time.Millisecond):
		log.Printf("Email queue overloaded for %s", email)
		return "", fmt.Errorf("email queue overloaded")
	}
}

func (s *emailService) sendVerificationEmail(
	email, username, code string,
	smtpConfig mail.EmailConfig,
	templateConfig mail.EmailTemplateConfig,
	appConfig mail.AppConfig,
) error {
	ticketNumber := fmt.Sprintf("CNF-%d-%d", time.Now().Unix()%10000, randInt(100, 999))
	log.Printf("Preparing email for %s with code %s and ticket %s", email, code, ticketNumber)

	data := struct {
		Username         string
		VerificationCode string
		SecurityURL      string
		SupportURL       string
		PrivacyURL       string
		TermsURL         string
	}{
		Username:         username,
		VerificationCode: code,
		SecurityURL:      appConfig.SecurityURL,
		SupportURL:       appConfig.SupportURL,
		PrivacyURL:       appConfig.PrivacyURL,
		TermsURL:         appConfig.TermsURL,
	}

	// Template processing
	log.Printf("Loading template from: %s", templateConfig.VerificationTemplatePath)
	templateContent, err := os.ReadFile(templateConfig.VerificationTemplatePath)
	if err != nil {
		log.Printf("Failed to read template: %v", err)
		return fmt.Errorf("failed to read template: %w", err)
	}
	log.Printf("Template content length: %d bytes", len(templateContent))

	var buf bytes.Buffer
	tmpl, err := template.New("verificationEmail").Parse(string(templateContent))
	if err != nil {
		log.Printf("Template parsing failed: %v", err)
		return fmt.Errorf("template parsing failed: %w", err)
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("Template execution failed: %v", err)
		return fmt.Errorf("template execution failed: %w", err)
	}
	log.Printf("Rendered email length: %d bytes", buf.Len())

	// Email sending
	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Confirm Your Ticket Zetu Account")
	m.SetBody("text/html", buf.String())

	log.Printf("Sending email to %s from %s via %s:%d", email, smtpConfig.FromEmail, smtpConfig.SMTPHost, smtpConfig.SMTPPort)
	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email to %s: %v", email, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("Email sent successfully to %s", email)
	return nil
}
