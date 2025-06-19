package mail_service

import (
	"bytes"
	"errors"
	"html/template"
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
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New("email queue overloaded")
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

	data := struct {
		Username       string
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
		DeviceInfo:     deviceInfo,
		IPAddress:      ipAddress,
		LoginTime:      loginTime.Format("2006-01-02 15:04:05"),
		WarningMessage: warningMessage,
		SecurityURL:    appConfig.SecurityURL,
		SupportURL:     appConfig.SupportURL,
		PrivacyURL:     appConfig.PrivacyURL,
		TermsURL:       appConfig.TermsURL,
	}

	templateContent, err := os.ReadFile(templateConfig.LoginWarningTemplatePath)
	if err != nil {
		return errors.New("failed to read template")
	}

	var buf bytes.Buffer
	tmpl, err := template.New("loginWarningEmail").Parse(string(templateContent))
	if err != nil {
		return errors.New("template parsing failed")
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return errors.New("template execution failed")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpConfig.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", buf.String())

	d := gomail.NewDialer(smtpConfig.SMTPHost, smtpConfig.SMTPPort, smtpConfig.SMTPUsername, smtpConfig.SMTPPassword)
	if err := d.DialAndSend(m); err != nil {
		return errors.New("failed to send email")
	}
	return nil
}
