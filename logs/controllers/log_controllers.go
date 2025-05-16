package controller

import (
	"fmt"
	"strconv"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/logs/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LogController struct {
	Service *service.LogService
}

// NewLogController creates a new LogController
func NewLogController(service *service.LogService) *LogController {
	return &LogController{Service: service}
}

// GetLogs retrieves logs based on optional query filters.
func (lc *LogController) GetLogs(c *fiber.Ctx, logHandler *handler.LogHandler) error {
	query := c.Queries()
	conditions := make(map[string]interface{})
	var args []interface{}

	// Filter by ip_address
	if ip := query["ip_address"]; ip != "" {
		conditions["ip_address = ?"] = ip
		args = append(args, ip)
	}

	// Filter by route
	if route := query["route"]; route != "" {
		conditions["route = ?"] = route
		args = append(args, route)
	}

	// Filter by message (LIKE for partial matches)
	if message := query["message"]; message != "" {
		conditions["message LIKE ?"] = "%" + message + "%"
		args = append(args, "%"+message+"%")
	}

	// Filter by level
	if level := query["level"]; level != "" {
		conditions["level = ?"] = level
		args = append(args, level)
	}

	// Filter by date (YYYY-MM-DD)
	if date := query["date"]; date != "" {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return logHandler.LogError(c, fmt.Errorf("invalid date format, use YYYY-MM-DD"), fiber.StatusBadRequest)
		}
		start := parsedDate.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		conditions["created_at >= ? AND created_at < ?"] = []interface{}{start, end}
		args = append(args, start, end)
	}

	// Filter by month (YYYY-MM)
	if month := query["month"]; month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return logHandler.LogError(c, fmt.Errorf("invalid month format, use YYYY-MM"), fiber.StatusBadRequest)
		}
		start := parsedMonth
		end := parsedMonth.AddDate(0, 1, 0)
		conditions["created_at >= ? AND created_at < ?"] = []interface{}{start, end}
		args = append(args, start, end)
	}

	// Pagination
	limit, _ := strconv.Atoi(query["limit"])
	if limit <= 0 {
		limit = 100 // Default limit
	}
	offset, _ := strconv.Atoi(query["offset"])
	if offset < 0 {
		offset = 0
	}

	logs, err := lc.Service.GetLogs(conditions, args, limit, offset)
	if err != nil {
		return logHandler.LogError(c, fmt.Errorf("failed to retrieve logs: %v", err), fiber.StatusInternalServerError)
	}

	return logHandler.LogSuccess(c, logs, "Logs retrieved successfully", true)
}

// DeleteLogs deletes logs based on optional query filters.

func (lc *LogController) DeleteLogs(c *fiber.Ctx, logHandler *handler.LogHandler) error {
	query := c.Queries()
	conditions := make(map[string]interface{})
	var args []interface{}

	// Filter by ip_address
	if ip := query["ip_address"]; ip != "" {
		conditions["ip_address = ?"] = ip
		args = append(args, ip)
	}

	// Filter by route
	if route := query["route"]; route != "" {
		conditions["route = ?"] = route
		args = append(args, route)
	}

	// Filter by message (LIKE for partial matches)
	if message := query["message"]; message != "" {
		conditions["message LIKE ?"] = "%" + message + "%"
		args = append(args, "%"+message+"%")
	}

	// Filter by level
	if level := query["level"]; level != "" {
		conditions["level = ?"] = level
		args = append(args, level)
	}

	// Filter by date (YYYY-MM-DD)
	if date := query["date"]; date != "" {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return logHandler.LogError(c, fmt.Errorf("invalid date format, use YYYY-MM-DD"), fiber.StatusBadRequest)
		}
		start := parsedDate.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		conditions["created_at >= ? AND created_at < ?"] = []interface{}{start, end}
		args = append(args, start, end)
	}

	// Filter by month (YYYY-MM)
	if month := query["month"]; month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return logHandler.LogError(c, fmt.Errorf("invalid month format, use YYYY-MM"), fiber.StatusBadRequest)
		}
		start := parsedMonth
		end := parsedMonth.AddDate(0, 1, 0)
		conditions["created_at >= ? AND created_at < ?"] = []interface{}{start, end}
		args = append(args, start, end)
	}

	count, err := lc.Service.DeleteLogs(conditions, args)
	if err != nil {
		return logHandler.LogError(c, fmt.Errorf("failed to delete logs: %v", err), fiber.StatusInternalServerError)
	}

	return logHandler.LogSuccess(c, nil, fmt.Sprintf("Deleted %d logs", count), true)
}
