package authentication

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	model "ticket-zetu-api/modules/users/models/authorization"
	"ticket-zetu-api/modules/users/models/members"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/gorm"
)

// SignUpRequest defines the structure for signup requests
type SignUpRequest struct {
	Username    string `json:"username" validate:"required,min=3,max=50,alphanum"`
	FirstName   string `json:"first_name" validate:"required,min=2,max=100"`
	LastName    string `json:"last_name" validate:"required,min=2,max=100"`
	Email       string `json:"email" validate:"required,email,max=255"`
	Phone       string `json:"phone" validate:"required,min=10,max=20"`
	Password    string `json:"password" validate:"required,min=8"`
	DateOfBirth string `json:"date_of_birth" validate:"omitempty,datetime=2006-01-02"`
}

// SignUp handles user registration
func (c *AuthController) SignUp(ctx *fiber.Ctx) error {
	var req SignUpRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("invalid request payload: "+err.Error()), fiber.StatusBadRequest)
	}

	// Validate input
	if err := c.validator.Struct(req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

	// Validate username and email uniqueness
	if err := c.userService.ValidateUserExists(req.Username, req.Email); err != nil {
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	// Parse date_of_birth if provided
	var dob *time.Time
	if req.DateOfBirth != "" {
		parsedDob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			return c.logHandler.LogError(ctx, errors.New("invalid date_of_birth format, use YYYY-MM-DD"), fiber.StatusBadRequest)
		}
		if !isUserOver16(parsedDob) {
			return c.logHandler.LogError(ctx, errors.New("user must be at least 16 years old"), fiber.StatusBadRequest)
		}
		dob = &parsedDob
	}

	// Fetch the default "user" role
	var role model.Role
	if err := c.db.Where("role_name = ?", "guest").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, errors.New("default guest role not found"), fiber.StatusInternalServerError)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	// Hash password with Argon2id
	userID := uuid.New().String()
	hashed := argon2.IDKey([]byte(req.Password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encodedHash := base64.RawStdEncoding.EncodeToString(hashed)
	fmt.Println(req)

	// Create user and related entities
	user := members.User{
		ID:          userID,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		DateOfBirth: dob,
		CreatedBy:   "system",
		RoleID:      role.ID,
	}

	prefs := members.UserPreferences{
		UserID:   uuid.MustParse(user.ID),
		Language: "en",
		Theme:    "light",
		Timezone: "UTC",
	}

	// Perform database operations in a transaction
	tx := c.db.Begin()
	if tx.Error != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Use userService.CreateUser to handle user creation
	if err := c.userService.CreateUser(tx, &user, encodedHash, &prefs, nil); err != nil {
		tx.Rollback()
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	if err := tx.Commit().Error; err != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
	}

	// Return only success message
	return c.logHandler.LogSuccess(ctx, nil, "Account created successfully", true)
}
