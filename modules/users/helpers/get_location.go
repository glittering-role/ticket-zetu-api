package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ticket-zetu-api/logs/handler"

	"github.com/gofiber/fiber/v2"
)

type Location struct {
	IPAddress string `json:"ip_address"`
	City      string `json:"city"`
	State     string `json:"region"`
	Country   string `json:"country"`
	Continent string `json:"continent"`
	Zip       string `json:"postal"`
	Timezone  string `json:"timezone"`
}

type GeolocationService struct {
	logHandler *handler.LogHandler
	apiToken   string
	client     *http.Client
}

// NewGeolocationService creates a new instance of GeolocationService
func NewGeolocationService(logHandler *handler.LogHandler, apiToken string) *GeolocationService {
	return &GeolocationService{
		logHandler: logHandler,
		apiToken:   apiToken,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GeolocationMiddleware fetches user location based on IP and stores it in context
func (s *GeolocationService) GeolocationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := getClientIP(c)
		if s.logHandler != nil {
			s.logHandler.LogWarning(c, fmt.Sprintf("Invalid or local IP detected: %s", ip), fiber.StatusBadRequest, nil)
		}

		location := &Location{
			IPAddress: ip,
			Country:   "Unknown",
			Continent: "Unknown",
			Timezone:  "UTC",
		}
		c.Locals("user_location", location)

		// Create a context with timeout for the async API call
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		defer cancel()

		// Channel to receive location data
		resultCh := make(chan *Location, 1)

		// Perform API call asynchronously
		go func() {
			defer close(resultCh)

			const maxRetries = 2
			for attempt := 1; attempt <= maxRetries; attempt++ {
				url := fmt.Sprintf("https://ipinfo.io/%s/json?token=%s", ip, s.apiToken)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				if err != nil {
					if s.logHandler != nil {
						s.logHandler.LogError(c, fmt.Errorf("attempt %d: failed to create ipinfo.io request: %w", attempt, err), fiber.StatusInternalServerError)
					}
					return
				}

				resp, err := s.client.Do(req)
				if err != nil {
					if s.logHandler != nil {
						s.logHandler.LogError(c, fmt.Errorf("attempt %d: failed to fetch location from ipinfo.io: %w", attempt, err), fiber.StatusInternalServerError)
					}
					if attempt < maxRetries {
						time.Sleep(time.Duration(100*attempt) * time.Millisecond)
						continue
					}
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					if s.logHandler != nil {
						s.logHandler.LogError(c, fmt.Errorf("attempt %d: ipinfo.io returned status %d: %s, body: %s", attempt, resp.StatusCode, resp.Status, string(body)), fiber.StatusInternalServerError)
					}
					if attempt < maxRetries {
						time.Sleep(time.Duration(100*attempt) * time.Millisecond)
						continue
					}
					return
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					if s.logHandler != nil {
						s.logHandler.LogError(c, fmt.Errorf("attempt %d: failed to read ipinfo.io response: %w", attempt, err), fiber.StatusInternalServerError)
					}
					if attempt < maxRetries {
						time.Sleep(time.Duration(100*attempt) * time.Millisecond)
						continue
					}
					return
				}

				var data map[string]interface{}
				if err := json.Unmarshal(body, &data); err != nil {
					if s.logHandler != nil {
						s.logHandler.LogError(c, fmt.Errorf("attempt %d: failed to parse ipinfo.io response: %w, body: %s", attempt, err, string(body)), fiber.StatusInternalServerError)
					}
					if attempt < maxRetries {
						time.Sleep(time.Duration(100*attempt) * time.Millisecond)
						continue
					}
					return
				}

				fetchedLocation := &Location{
					IPAddress: ip,
					City:      getString(data, "city"),
					State:     getString(data, "region"),
					Country:   getString(data, "country"),
					Continent: getString(data, "continent"),
					Zip:       getString(data, "postal"),
					Timezone:  getString(data, "timezone"),
				}

				// Derive continent from timezone if not provided
				if fetchedLocation.Continent == "" && fetchedLocation.Timezone != "" {
					parts := strings.Split(fetchedLocation.Timezone, "/")
					if len(parts) > 0 && parts[0] != "" {
						fetchedLocation.Continent = parts[0]
					} else {
						fetchedLocation.Continent = "Unknown"
					}
				}

				select {
				case resultCh <- fetchedLocation:
					return
				case <-ctx.Done():
					if s.logHandler != nil {
						s.logHandler.LogWarning(c, fmt.Sprintf("ipinfo.io attempt %d timed out or context cancelled", attempt), fiber.StatusBadRequest, nil)
					}
					return
				}
			}
		}()

		// Wait for result or timeout
		select {
		case fetchedLocation := <-resultCh:
			if fetchedLocation != nil {
				location = fetchedLocation
				c.Locals("user_location", location)
			}
		case <-ctx.Done():
			if s.logHandler != nil {
				s.logHandler.LogWarning(c, "Timeout waiting for ipinfo.io response, using default location", fiber.StatusBadRequest, nil)
			}
		}

		return c.Next()
	}
}

// getClientIP retrieves the client IP, prioritizing X-Forwarded-For or X-Real-IP headers
func getClientIP(c *fiber.Ctx) string {
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	// Fallback to X-Real-IP header
	if realIP := c.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}
	return c.IP()
}

// getString safely extracts string from map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok && val != nil {
		return val.(string)
	}
	return ""
}
