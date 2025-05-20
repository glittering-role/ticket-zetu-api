package service

import (
	"context"
	"fmt"
	"log"
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
}

func NewLogService(db *gorm.DB, bufferSize int, flushPeriod time.Duration, env string) *LogService {
	ctx, cancel := context.WithCancel(context.Background())
	service := &LogService{
		db:          db,
		logChan:     make(chan model.Log, bufferSize),
		bufferSize:  bufferSize,
		flushPeriod: flushPeriod,
		ctx:         ctx,
		cancel:      cancel,
		Env:         env,
	}

	service.wg.Add(1)
	go service.processLogs()

	return service
}

// Log queues a log entry
func (s *LogService) Log(logEntry model.Log) {
	// Only set environment if not set
	if logEntry.Environment == nil {
		logEntry.Environment = &s.Env
	}

	select {
	case s.logChan <- logEntry:
	default:
		log.Printf("Log channel full, dropped entry: %v", logEntry.Message)
	}
}

// processLogs processes log entries from the channel
func (s *LogService) processLogs() {
	defer s.wg.Done()

	var batch []model.Log
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			if len(batch) > 0 {
				s.flushLogs(batch)
			}
			return

		case entry := <-s.logChan:
			// Try to find existing log to increment occurrences
			existing, err := s.findExistingLog(entry)
			if err == nil && existing != nil {
				// Successfully found an existing log, increment its occurrences
				s.incrementOccurrence(existing, entry)
			} else {
				// No existing log found or error occurred, add to batch for insertion
				batch = append(batch, entry)
				if len(batch) >= s.bufferSize {
					s.flushLogs(batch)
					batch = make([]model.Log, 0, s.bufferSize) // Reset with capacity
				}
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.flushLogs(batch)
				batch = make([]model.Log, 0, s.bufferSize) // Reset with capacity
			}
		}
	}
}

// findExistingLog checks for an existing log entry with the same IP, route, and message
func (s *LogService) findExistingLog(entry model.Log) (*model.Log, error) {
	if entry.IPAddress == nil || entry.Route == nil {
		log.Printf("Skipping deduplication: missing IPAddress or Route")
		return nil, nil
	}

	var existing model.Log
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	// Use a transaction for consistency
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Add level to the query for more specific matching
		result := tx.Where(
			"ip_address = ? AND route = ? AND message = ? AND level = ? AND created_at >= ?",
			*entry.IPAddress, *entry.Route, entry.Message, entry.Level, oneHourAgo,
		).Order("created_at DESC").First(&existing)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return gorm.ErrRecordNotFound
			}
			return fmt.Errorf("database error: %w", result.Error)
		}

		return nil
	})

	if err == gorm.ErrRecordNotFound {
		log.Printf("No existing log found for IP: %s, Route: %s, Message: %s",
			*entry.IPAddress, *entry.Route, entry.Message)
		return nil, nil
	}

	if err != nil {
		log.Printf("Error finding existing log: %v", err)
		return nil, fmt.Errorf("failed to find existing log: %w", err)
	}

	log.Printf("Found existing log ID: %d for IP: %s, Route: %s, Message: %s",
		existing.ID, *entry.IPAddress, *entry.Route, entry.Message)
	return &existing, nil
}

// incrementOccurrence updates the occurrence count for an existing log
func (s *LogService) incrementOccurrence(existing *model.Log, newEntry model.Log) {
	// Prepare update data
	update := map[string]interface{}{
		"occurrences": gorm.Expr("occurrences + 1"),
		"updated_at":  time.Now(),
	}

	// Update context, stack, and other fields if provided
	if newEntry.Context != nil {
		update["context"] = newEntry.Context
	}
	if newEntry.Stack != nil {
		update["stack"] = newEntry.Stack
	}
	if newEntry.StatusCode != nil {
		update["status_code"] = newEntry.StatusCode
	}
	if newEntry.Method != nil {
		update["method"] = newEntry.Method
	}
	if newEntry.UserAgent != nil {
		update["user_agent"] = newEntry.UserAgent
	}
	if newEntry.File != nil {
		update["file"] = newEntry.File
	}
	if newEntry.Line != nil {
		update["line"] = newEntry.Line
	}
	if newEntry.Level != "" {
		update["level"] = newEntry.Level
	}

	// Execute the update with a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Log{}).
			Where("id = ?", existing.ID).
			Updates(update)

		if result.Error != nil {
			return result.Error
		}

		// Check if the update was actually applied
		if result.RowsAffected == 0 {
			return fmt.Errorf("no rows affected when updating log ID: %d", existing.ID)
		}

		return nil
	})

	if err != nil {
		log.Printf("Failed to increment occurrence for log ID: %d, Error: %v", existing.ID, err)
	} else {
		log.Printf("Incremented occurrences for log ID: %d, New count: %d",
			existing.ID, existing.Occurrences+1)
	}
}

// flushLogs saves a batch of logs to the database
func (s *LogService) flushLogs(logs []model.Log) {
	if len(logs) == 0 {
		return
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&logs).Error
	})
	if err != nil {
		log.Printf("Failed to save logs: %v", err)
	}
}

// GetLogs retrieves logs based on conditions
func (s *LogService) GetLogs(conditions map[string]interface{}, args []interface{}, limit, offset int) ([]model.Log, error) {
	var logs []model.Log
	query := s.db.Model(&model.Log{}).Unscoped() // Include soft-deleted logs
	for condition, value := range conditions {
		log.Printf("Applying condition: %s with value: %v", condition, value)
		query = query.Where(condition, value)
	}
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		log.Printf("Failed to retrieve logs: %v", err)
		return nil, fmt.Errorf("failed to retrieve logs: %w", err)
	}
	log.Printf("Retrieved %d logs with limit=%d, offset=%d, conditions=%v", len(logs), limit, offset, conditions)
	return logs, nil
}

// DeleteLogs deletes logs based on conditions
func (s *LogService) DeleteLogs(conditions map[string]interface{}, args []interface{}) (int64, error) {
	query := s.db.Model(&model.Log{}).Unscoped() // Include soft-deleted logs
	for condition, value := range conditions {
		log.Printf("Applying condition: %s with value: %v", condition, value)
		query = query.Where(condition, value)
	}
	result := query.Delete(&model.Log{})
	if result.Error != nil {
		log.Printf("Failed to delete logs: %v", result.Error)
		return 0, fmt.Errorf("failed to delete logs: %w", result.Error)
	}
	log.Printf("Deleted %d logs", result.RowsAffected)
	return result.RowsAffected, nil
}

// Shutdown gracefully stops the log service
func (s *LogService) Shutdown() {
	s.cancel()
	s.wg.Wait()
	close(s.logChan)
}
