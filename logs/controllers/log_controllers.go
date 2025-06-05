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

// Logs docs
// @Summary Get Logs
// @Description Retrieve logs with optional filters
// @Tags Logs
// @Accept json
// @Produce json
// @Param ip_address query string false "Filter by IP address"
// @Param route query string false "Filter by route"
// @Param message query string false "Filter by message (partial match)"
// @Param level query string false "Filter by log level (e.g., info, error)"
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param month query string false "Filter by month (YYYY-MM)"
// @Param limit query int false "Number of logs to retrieve (default 100)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {object} map[string]interface{} "Logs retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs [get]
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

// DeleteLogs docs
// @Summary Delete Logs
// @Description Delete logs with optional filters
// @Tags Logs
// @Accept json
// @Produce json
// @Param ip_address query string false "Filter by IP address"
// @Param route query string false "Filter by route"
// @Param message query string false "Filter by message (partial match)"
// @Param level query string false "Filter by log level (e.g., info, error)"
// @Param date query string false "Filter by date (YYYY-MM-DD)"
// @Param month query string false "Filter by month (YYYY-MM)"
// @Success 200 {object} map[string]interface{} "Logs deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /logs [delete]
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
