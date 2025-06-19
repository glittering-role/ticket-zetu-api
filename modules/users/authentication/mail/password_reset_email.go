package mail_service

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
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
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New("email queue overloaded")
	}
}

func (s *emailService) sendPasswordResetEmail(
	email, username, resetToken string,
	smtpConfig mail.EmailConfig,
	templateConfig mail.EmailTemplateConfig,
	appConfig mail.AppConfig,
) error {
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
		Username:   username,
		ResetURL:   resetURL,
		SupportURL: appConfig.SupportURL,
		PrivacyURL: appConfig.PrivacyURL,
		TermsURL:   appConfig.TermsURL,
		ExpiryTime: time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	templateContent, err := os.ReadFile(templateConfig.PasswordResetTemplatePath)
	if err != nil {
		return errors.New("failed to read template")
	}

	var buf bytes.Buffer
	tmpl, err := template.New("passwordResetEmail").Parse(string(templateContent))
	if err != nil {
		return errors.New("template parsing failed")
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return errors.New("template execution failed")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Reset Your Ticket Zetu Password")
	m.SetBody("text/html", buf.String())

	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		return errors.New("failed to send email")
	}
	return nil
}
