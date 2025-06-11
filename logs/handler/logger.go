package handler

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"ticket-zetu-api/logs/model"
	"ticket-zetu-api/logs/service"
	response "ticket-zetu-api/utils/res"

	"github.com/gofiber/fiber/v2"
)

type LogHandler struct {
	Service *service.LogService
}

// LogError logs an error with a stack trace and returns a response
func (lh *LogHandler) LogError(c *fiber.Ctx, err error, statusCode int, data ...interface{}) error {
	log := lh.createBaseLog(c, "error", err.Error(), statusCode)
	log.Stack = lh.getStackTrace() // Capture stack trace for errors

	// Log in background
	go lh.Service.Log(log)

	var responseData interface{}
	if len(data) > 0 {
		responseData = data[0]
	}
	return response.Error(c, statusCode, err.Error(), responseData)
}

// LogInfo logs a general informational message without sending a response
func (lh *LogHandler) LogInfo(c *fiber.Ctx, message string) {
	log := lh.createBaseLog(c, "info", message, fiber.StatusOK)
	go lh.Service.Log(log)
}

// LogSuccess logs a success message if logInfo is true and returns a response
func (lh *LogHandler) LogSuccess(c *fiber.Ctx, data interface{}, message string, logInfo ...bool) error {
	if len(logInfo) == 0 || !logInfo[0] {
		return response.Success(c, message, data)
	}

	log := lh.createBaseLog(c, "info", message, fiber.StatusOK)

	// Log in background
	go lh.Service.Log(log)

	return response.Success(c, message, data)
}

// LogWarning logs a warning and returns a response
func (lh *LogHandler) LogWarning(c *fiber.Ctx, message string, statusCode int, data interface{}) error {
	log := lh.createBaseLog(c, "warning", message, statusCode)

	// Log in background
	go lh.Service.Log(log)

	return response.Warning(c, statusCode, message, data)
}

// createBaseLog constructs a common log structure with comprehensive context
func (lh *LogHandler) createBaseLog(c *fiber.Ctx, level, message string, statusCode int) model.Log {
	file, line := lh.getCallerInfo()
	path := c.Path()
	method := c.Method()
	env := lh.Service.Env

	// Use the improved IP address resolution method
	ip := lh.getClientIP(c)

	userAgent := c.Get("User-Agent")
	queryString := c.Context().URI().QueryArgs().String()

	// Build context JSON with request details
	contextMap := map[string]interface{}{
		"query":   queryString,
		"body":    string(c.Body()),  // Include request body
		"params":  c.AllParams(),     // Include path parameters
		"headers": c.GetReqHeaders(), // Include headers
	}
	contextJSON, err := json.Marshal(contextMap)
	if err != nil {
		contextJSON = []byte(`{"error":"failed to marshal context"}`)
	}
	context := string(contextJSON)

	return model.Log{
		Level:       level,
		Message:     message,
		Route:       &path,
		Method:      &method,
		StatusCode:  &statusCode,
		File:        &file,
		Line:        &line,
		Environment: &env,
		IPAddress:   &ip,
		UserAgent:   &userAgent,
		Context:     &context,
	}
}

// getClientIP extracts the client IP address considering proxy headers
func (lh *LogHandler) getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first (most common proxy header)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// The X-Forwarded-For header can contain multiple IPs
		// The leftmost IP is usually the original client IP
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check other common proxy headers
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := c.Get("CF-Connecting-IP"); ip != "" { // Cloudflare
		return ip
	}

	if ip := c.Get("True-Client-IP"); ip != "" {
		return ip
	}

	// Fallback to Fiber's IP method (which typically returns the direct connection IP)
	return c.IP()
}

// getCallerInfo retrieves the file and line number of the caller
func (lh *LogHandler) getCallerInfo() (string, int) {
	skip := 3 // Start deeper to skip LogHandler methods
	for {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			return "unknown", 0
		}
		// Skip internal Fiber, runtime, and handler frames
		if !strings.Contains(file, "github.com/gofiber/fiber") &&
			!strings.Contains(file, "runtime/") &&
			!strings.Contains(file, "ticket-zetu-api/logs/handler") {
			// Trim path to last 3 segments
			parts := strings.Split(file, "/")
			if len(parts) > 3 {
				file = strings.Join(parts[len(parts)-3:], "/")
			}
			return file, line
		}
		skip++
		if skip > 20 { // Prevent infinite loop
			return "unknown", 0
		}
	}
}

// getStackTrace captures a filtered stack trace for errors
func (lh *LogHandler) getStackTrace() *string {
	var stackBuilder strings.Builder
	skip := 3 // Start deeper to capture caller context
	for {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		// Include only relevant frames
		if !strings.Contains(file, "github.com/gofiber/fiber") &&
			!strings.Contains(file, "runtime/") &&
			!strings.Contains(file, "ticket-zetu-api/logs/handler") {
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				parts := strings.Split(file, "/")
				if len(parts) > 3 {
					file = strings.Join(parts[len(parts)-3:], "/")
				}
				stackBuilder.WriteString(fmt.Sprintf("%s:%d %s\n", file, line, fn.Name()))
			}
		}
		skip++
		if skip > 20 { // Limit stack depth
			break
		}
	}
	stack := stackBuilder.String()
	if stack == "" {
		return nil
	}
	return &stack
}
