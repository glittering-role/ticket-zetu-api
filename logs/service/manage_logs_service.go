package service

import (
	"fmt"
	"log"
	"ticket-zetu-api/logs/model"
)

// GetLogs retrieves logs based on conditions
func (s *LogService) GetLogs(conditions map[string]interface{}, args []interface{}, limit, offset int) ([]model.Log, error) {
	var logs []model.Log
	query := s.db.Model(&model.Log{}).Unscoped()
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
	query := s.db.Model(&model.Log{}).Unscoped()
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
