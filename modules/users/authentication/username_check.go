package authentication

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"ticket-zetu-api/logs/handler"
	"ticket-zetu-api/modules/users/models/members"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UsernameCheck struct {
	db          *gorm.DB
	logHandler  *handler.LogHandler
	cache       map[string]UsernameCacheEntry
	cacheMutex  sync.RWMutex
	rateLimiter map[string]time.Time
}

type UsernameCacheEntry struct {
	Response  UsernameCheckResponse
	ExpiresAt time.Time
}

func CheckUsernameAvailability(db *gorm.DB, logHandler *handler.LogHandler) *UsernameCheck {
	return &UsernameCheck{
		db:          db,
		logHandler:  logHandler,
		cache:       make(map[string]UsernameCacheEntry),
		rateLimiter: make(map[string]time.Time),
	}
}

type UsernameCheckRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
}

type UsernameCheckResponse struct {
	Available   bool     `json:"available"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// International username standards:
// 1. 3-50 characters
// 2. Only alphanumeric, underscore, hyphen, and dot
// 3. No consecutive special characters
// 4. Must start and end with alphanumeric
// 5. No offensive words
var (
	usernameRegex    = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9_\-.]*[a-zA-Z0-9])?$`)
	consecutiveRegex = regexp.MustCompile(`[_\-.]{2,}`)
	blacklist        = []string{"admin", "root", "moderator"}
)

func (uc *UsernameCheck) CheckUsername(ctx *fiber.Ctx) error {
	// Rate limiting by IP
	ip := ctx.IP()
	if lastRequest, exists := uc.rateLimiter[ip]; exists && time.Since(lastRequest) < 1*time.Second {
		return uc.logHandler.LogError(ctx, errors.New("too many requests"), fiber.StatusTooManyRequests)
	}
	uc.rateLimiter[ip] = time.Now()

	// Get username from query parameter
	username := strings.TrimSpace(ctx.Query("username"))
	if username == "" {
		return uc.logHandler.LogError(ctx, errors.New("username parameter is required"), fiber.StatusBadRequest)
	}

	// Check cache first
	uc.cacheMutex.RLock()
	if cached, exists := uc.cache[username]; exists && time.Now().Before(cached.ExpiresAt) {
		uc.cacheMutex.RUnlock()
		return uc.logHandler.LogSuccess(ctx, cached.Response, "Username check completed (cached)", false)
	}
	uc.cacheMutex.RUnlock()

	// Validate against international standards
	if err := validateInternationalUsername(username); err != nil {
		return uc.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Check username availability
	var user members.User
	err := uc.db.Where("username = ?", username).First(&user).Error
	response := UsernameCheckResponse{
		Available:   true,
		Message:     "Username is available",
		Suggestions: []string{},
	}

	if err == nil {
		response.Available = false
		response.Message = "Username is already taken"
		response.Suggestions = uc.generateUsernameSuggestions(username)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uc.logHandler.LogError(ctx, errors.New("failed to check username availability"), fiber.StatusInternalServerError)
	}

	// Cache the result
	uc.cacheMutex.Lock()
	uc.cache[username] = UsernameCacheEntry{
		Response:  response,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	uc.cacheMutex.Unlock()

	return uc.logHandler.LogSuccess(ctx, response, "Username check completed", false)
}

func validateInternationalUsername(username string) error {
	// Length check
	if len(username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	if len(username) > 50 {
		return errors.New("username cannot exceed 50 characters")
	}

	// Character set check
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain letters, numbers, underscores, hyphens, and dots")
	}

	// No consecutive special characters
	if consecutiveRegex.MatchString(username) {
		return errors.New("username cannot contain consecutive special characters")
	}

	// No blacklisted words
	lowerUsername := strings.ToLower(username)
	for _, word := range blacklist {
		if strings.Contains(lowerUsername, word) {
			return errors.New("username contains restricted words")
		}
	}

	// No reserved prefixes/suffixes
	if strings.HasPrefix(lowerUsername, "sys_") || strings.HasSuffix(lowerUsername, "_system") {
		return errors.New("username uses reserved prefixes/suffixes")
	}

	return nil
}

func (uc *UsernameCheck) generateUsernameSuggestions(username string) []string {
	suggestions := []string{}
	baseUsername := strings.ToLower(username)

	// More sophisticated suggestion patterns
	suggestionPatterns := []struct {
		pattern  string
		maxTries int
	}{
		{"%s%d", 5},   // username1, username2, etc.
		{"%s_%d", 5},  // username_1, username_2, etc.
		{"%s%s", 3},   // Add random suffix
		{"the_%s", 1}, // the_username
		{"real%s", 1}, // realusername
		{"%s_go", 1},  // username_go
		{"%s_now", 1}, // username_now
		{"%s_%s", 3},  // username_xyz
	}

	// Generate suggestions
	for _, pattern := range suggestionPatterns {
		for i := 1; i <= pattern.maxTries; i++ {
			suggestion := ""
			switch {
			case strings.Contains(pattern.pattern, "%d"):
				suggestion = fmt.Sprintf(pattern.pattern, baseUsername, i)
			case strings.Count(pattern.pattern, "%s") == 2:
				suggestion = fmt.Sprintf(pattern.pattern, baseUsername, randomString(3))
			default:
				suggestion = fmt.Sprintf(pattern.pattern, baseUsername)
			}

			// Verify suggestion availability
			var user members.User
			if err := uc.db.Where("username = ?", suggestion).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				suggestions = append(suggestions, suggestion)
				if len(suggestions) >= 5 {
					return suggestions[:5] // Return early if we have enough
				}
			}
		}
	}

	return suggestions
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
