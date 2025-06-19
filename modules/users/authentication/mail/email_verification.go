package mail_service

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"html/template"
	"math/big"
	"os"
	"time"

	"ticket-zetu-api/mail"

	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

func (s *emailService) generateVerificationCode() (string, error) {
	// Generate a random number between 0 and 99999999 (8 digits)
	max := big.NewInt(100000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate random number")
	}
	code := fmt.Sprintf("%08d", n.Int64())
	if len(code) != 8 {
		return s.generateVerificationCode()
	}
	return code, nil
}

// GenerateAndSendVerificationCode generates and queues an email verification code
func (s *emailService) GenerateAndSendVerificationCode(c *fiber.Ctx, email, username, userID string) (string, error) {
	code, err := s.generateVerificationCode()
	if err != nil {
		return "", errors.New("verification code generation failed")
	}

	// Validate code length
	if len(code) != 8 {
		return "", errors.New("generated code is not 8 digits")
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
		return code, nil
	case <-time.After(100 * time.Millisecond):
		return "", errors.New("email queue overloaded")
	}
}

func (s *emailService) sendVerificationEmail(
	email, username, code string,
	smtpConfig mail.EmailConfig,
	templateConfig mail.EmailTemplateConfig,
	appConfig mail.AppConfig,
) error {
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
	templateContent, err := os.ReadFile(templateConfig.VerificationTemplatePath)
	if err != nil {
		return errors.New("failed to read template")
	}

	var buf bytes.Buffer
	tmpl, err := template.New("verificationEmail").Parse(string(templateContent))
	if err != nil {
		return errors.New("template parsing failed")
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return errors.New("template execution failed")
	}

	// Email sending
	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Confirm Your Ticket Zetu Account")
	m.SetBody("text/html", buf.String())

	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		return errors.New("failed to send email")
	}
	return nil
}
