package helpers

import (
	"fmt"
	"strings"

	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
	"github.com/mssola/user_agent"
)

// DeviceInfo holds parsed device information
type DeviceInfo struct {
	Description string                    `json:"description"`
	Browser     string                    `json:"browser"`
	DeviceType  members.SessionDeviceType `json:"device_type"`
	OS          string                    `json:"os"`
}

type DeviceDetectionService struct {
	logHandler *handler.LogHandler
}

// NewDeviceDetectionService creates a new instance of DeviceDetectionService
func NewDeviceDetectionService(logHandler *handler.LogHandler) *DeviceDetectionService {
	return &DeviceDetectionService{
		logHandler: logHandler,
	}
}

// DeviceDetectionMiddleware parses User-Agent and stores device info in context
func (s *DeviceDetectionService) DeviceDetectionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			if s.logHandler != nil {
				s.logHandler.LogWarning(c, "Empty User-Agent detected", fiber.StatusBadRequest, nil)
			}
			userAgent = "Unknown"
		}

		ua := user_agent.New(userAgent)
		browser, _ := ua.Browser()
		os := ua.OS()
		var deviceType members.SessionDeviceType

		switch {
		case ua.Mobile():
			deviceType = members.DeviceMobile
		case strings.Contains(strings.ToLower(userAgent), "tablet"):
			deviceType = members.DeviceTablet
		case !ua.Mobile() && !strings.Contains(strings.ToLower(userAgent), "tablet"):
			deviceType = members.DeviceDesktop
		default:
			deviceType = members.DeviceUnknown
		}

		// Construct concise description
		browserDesc := browser
		if browserDesc == "" {
			browserDesc = "Unknown"
		}
		osDesc := os
		if osDesc == "" {
			osDesc = "Unknown"
		}
		var deviceDesc string
		switch deviceType {
		case members.DeviceMobile:
			deviceDesc = "Mobile"
		case members.DeviceTablet:
			deviceDesc = "Tablet"
		case members.DeviceDesktop:
			deviceDesc = "Desktop"
		default:
			deviceDesc = "Device"
		}
		description := fmt.Sprintf("%s %s on %s", browserDesc, deviceDesc, osDesc)

		deviceInfo := &DeviceInfo{
			Description: description,
			Browser:     browserDesc,
			DeviceType:  deviceType,
			OS:          osDesc,
		}

		c.Locals("device_info", deviceInfo)

		if s.logHandler != nil {
			logFields := fmt.Sprintf("Description: %s, Browser: %s, DeviceType: %s, OS: %s",
				description, browserDesc, deviceType, osDesc)
			s.logHandler.LogInfo(c, fmt.Sprintf("Detected device info: %s", logFields))
		}

		return c.Next()
	}
}
