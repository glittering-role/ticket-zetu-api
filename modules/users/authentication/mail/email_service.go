package mail_service

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/mail"

	"github.com/gofiber/fiber/v2"
)

// EmailService defines the interface for email operations
type EmailService interface {
	GenerateAndSendVerificationCode(c *fiber.Ctx, email, username, userID string) (string, error)
	SendLoginWarning(c *fiber.Ctx, email, username, userAgent, ipAddress string, loginTime time.Time, warningType string) error
	SendPasswordResetEmail(c *fiber.Ctx, email, username, resetToken string) error
	Shutdown()
}

// emailService holds dependencies
type emailService struct {
	logHandler  *handler.LogHandler
	config      *mail.Config
	jobQueue    chan emailJob
	workerGroup sync.WaitGroup
	shutdown    chan struct{}
}

type emailJob struct {
	execute func() error
	ctx     *fiber.Ctx
}

// NewEmailService creates a new email service instance
func NewEmailService(config *mail.Config, logHandler *handler.LogHandler, workers int) EmailService {
	service := &emailService{
		logHandler: logHandler,
		config:     config,
		jobQueue:   make(chan emailJob, 1000),
		shutdown:   make(chan struct{}),
	}

	// Start worker pool
	service.workerGroup.Add(workers)
	for i := 0; i < workers; i++ {
		go service.worker()
	}

	return service
}

func (s *emailService) worker() {
	defer s.workerGroup.Done()

	for {
		select {
		case job := <-s.jobQueue:
			if err := job.execute(); err != nil {
				s.logHandler.LogError(job.ctx, err, fiber.StatusInternalServerError)
			}
		case <-s.shutdown:
			return
		}
	}
}

func (s *emailService) Shutdown() {
	close(s.shutdown)
	s.workerGroup.Wait()
}

func (s *emailService) generateVerificationCode() (string, error) {
	// Generate a random number between 0 and 99999999 (8 digits)
	max := big.NewInt(100000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	code := fmt.Sprintf("%08d", n.Int64())
	if len(code) != 8 {
		// log.Printf("Generated code %s is not 8 digits, regenerating", code)
		return s.generateVerificationCode()
	}
	log.Printf("Generated verification code: %s", code)
	return code, nil
}

// randInt generates a random integer between min and max
func randInt(min, max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	return min + int(n.Int64())
}
