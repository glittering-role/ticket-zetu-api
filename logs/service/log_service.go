package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"ticket-zetu-api/logs/model"
	"time"

	"gorm.io/gorm"
)

type LogService struct {
	db          *gorm.DB
	logChan     chan model.Log
	bufferSize  int
	flushPeriod time.Duration
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	Env         string

	// Retry configuration
	maxRetries int
	retryDelay time.Duration

	// Dead-letter queue configuration
	dlqEnabled     bool
	dlqFile        string
	dlqMutex       sync.Mutex
	dlqBatchSize   int
	dlqFlushPeriod time.Duration
	dlqBuffer      []model.Log
}

func NewLogService(db *gorm.DB, bufferSize int, flushPeriod time.Duration, env string) *LogService {
	ctx, cancel := context.WithCancel(context.Background())
	service := &LogService{
		db:             db,
		logChan:        make(chan model.Log, bufferSize),
		bufferSize:     bufferSize,
		flushPeriod:    flushPeriod,
		ctx:            ctx,
		cancel:         cancel,
		Env:            env,
		maxRetries:     3,
		retryDelay:     1 * time.Second,
		dlqEnabled:     true,
		dlqFile:        "logs_dlq.json",
		dlqBatchSize:   100,
		dlqFlushPeriod: 5 * time.Minute,
		dlqBuffer:      make([]model.Log, 0, 100),
	}

	service.wg.Add(1)
	go service.processLogs()

	if service.dlqEnabled {
		service.wg.Add(1)
		go service.processDLQ()
	}

	return service
}

func (s *LogService) Log(logEntry model.Log) {
	if logEntry.Environment == nil {
		logEntry.Environment = &s.Env
	}

	select {
	case s.logChan <- logEntry:
	default:
		log.Printf("Log channel full, dropped entry: %v", logEntry.Message)
	}
}

func (s *LogService) processLogs() {
	defer s.wg.Done()

	var batch []model.Log
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			if len(batch) > 0 {
				s.flushLogsWithRetry(batch)
			}
			return

		case entry := <-s.logChan:
			batch = append(batch, entry)
			if len(batch) >= s.bufferSize {
				s.flushLogsWithRetry(batch)
				batch = make([]model.Log, 0, s.bufferSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushLogsWithRetry(batch)
				batch = make([]model.Log, 0, s.bufferSize)
			}
		}
	}
}

func (s *LogService) flushLogsWithRetry(logs []model.Log) {
	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		err := s.flushLogs(logs)
		if err == nil {
			return // Success
		}

		if attempt < s.maxRetries {
			log.Printf("Failed to flush logs (attempt %d/%d), retrying in %v: %v",
				attempt, s.maxRetries, s.retryDelay, err)
			time.Sleep(s.retryDelay)
			continue
		}

		log.Printf("Failed to flush logs after %d attempts: %v", s.maxRetries, err)
		s.handleFailedLogs(logs, err)
	}
}

func (s *LogService) flushLogs(logs []model.Log) error {
	if len(logs) == 0 {
		return nil
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&logs)
		if result.Error != nil {
			return result.Error
		}

		if int(result.RowsAffected) != len(logs) {
			return fmt.Errorf("only %d of %d logs were inserted",
				result.RowsAffected, len(logs))
		}
		return nil
	})
}

func (s *LogService) handleFailedLogs(logs []model.Log, err error) {
	if !s.dlqEnabled {
		log.Printf("DLQ disabled, dropping %d failed logs", len(logs))
		return
	}

	s.dlqMutex.Lock()
	defer s.dlqMutex.Unlock()

	// Append failed logs to DLQ buffer
	s.dlqBuffer = append(s.dlqBuffer, logs...)

	// If buffer exceeds batch size, flush to file
	if len(s.dlqBuffer) >= s.dlqBatchSize {
		if err := s.flushDLQ(); err != nil {
			log.Printf("Failed to flush DLQ: %v", err)
		}
	}
}

func (s *LogService) processDLQ() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.dlqFlushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.dlqMutex.Lock()
			if len(s.dlqBuffer) > 0 {
				if err := s.flushDLQ(); err != nil {
					log.Printf("Failed to flush DLQ on shutdown: %v", err)
				}
			}
			s.dlqMutex.Unlock()
			return

		case <-ticker.C:
			s.dlqMutex.Lock()
			if len(s.dlqBuffer) > 0 {
				if err := s.flushDLQ(); err != nil {
					log.Printf("Failed to flush DLQ: %v", err)
				}
			}
			s.dlqMutex.Unlock()
		}
	}
}

func (s *LogService) flushDLQ() error {
	if len(s.dlqBuffer) == 0 {
		return nil
	}

	file, err := os.OpenFile(s.dlqFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open DLQ file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, logEntry := range s.dlqBuffer {
		if err := encoder.Encode(logEntry); err != nil {
			return fmt.Errorf("failed to write log to DLQ: %w", err)
		}
	}

	log.Printf("Successfully wrote %d logs to DLQ file", len(s.dlqBuffer))
	s.dlqBuffer = s.dlqBuffer[:0] // Clear buffer
	return nil
}

func (s *LogService) RecoverFromDLQ() (int, error) {
	if !s.dlqEnabled {
		return 0, errors.New("DLQ is not enabled")
	}

	s.dlqMutex.Lock()
	defer s.dlqMutex.Unlock()

	file, err := os.Open(s.dlqFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No DLQ file exists
		}
		return 0, fmt.Errorf("failed to open DLQ file: %w", err)
	}
	defer file.Close()

	var logs []model.Log
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var logEntry model.Log
		if err := decoder.Decode(&logEntry); err != nil {
			return len(logs), fmt.Errorf("failed to decode log entry: %w", err)
		}
		logs = append(logs, logEntry)
	}

	if len(logs) == 0 {
		return 0, nil
	}

	// Attempt to re-insert logs
	successCount := 0
	for i := 0; i < len(logs); i += s.bufferSize {
		end := i + s.bufferSize
		if end > len(logs) {
			end = len(logs)
		}
		batch := logs[i:end]

		if err := s.flushLogs(batch); err != nil {
			// Put remaining logs back in DLQ
			s.dlqBuffer = append(logs[i:], s.dlqBuffer...)
			return successCount, fmt.Errorf("failed to recover logs from DLQ: %w", err)
		}
		successCount += len(batch)
	}

	// If all logs were recovered, delete the DLQ file
	if successCount == len(logs) {
		if err := os.Remove(s.dlqFile); err != nil {
			return successCount, fmt.Errorf("failed to remove DLQ file: %w", err)
		}
	}

	return successCount, nil
}
