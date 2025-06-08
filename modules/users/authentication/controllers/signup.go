package authentication

import (
	"encoding/base64"
	"errors"
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
	Username    string `json:"username" example:"johndoe99" validate:"required,min=3,max=50,alphanum"`
	FirstName   string `json:"first_name" example:"John" validate:"required,min=2,max=100"`
	LastName    string `json:"last_name" example:"Doe" validate:"required,min=2,max=100"`
	Email       string `json:"email" example:"john.doe@example.com" validate:"required,email,max=255"`
	Phone       string `json:"phone" example:"+12345678901" validate:"required,min=10,max=20"`
	Password    string `json:"password" example:"P@ssw0rd!" validate:"required,min=8"`
	DateOfBirth string `json:"date_of_birth,omitempty" example:"1990-05-20" validate:"omitempty,datetime=2006-01-02"`
}

// SignUp godoc
// @Summary Register a new user
// @Description Creates a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body authentication.SignUpRequest true "Signup data"
// @Success 201 {object} map[string]interface{} "Account created successfully"
// @Failure 400 {object} map[string]interface{} "Validation error"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/sign-up [post]
// SignUp handles user registration
func (c *AuthController) SignUp(ctx *fiber.Ctx) error {
	var req SignUpRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("invalid request payload: "+err.Error()), fiber.StatusBadRequest)
	}

	if err := c.validator.Struct(req); err != nil {
		return c.logHandler.LogError(ctx, errors.New("validation failed: "+err.Error()), fiber.StatusBadRequest)
	}

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

	var role model.Role
	if err := c.db.Where("role_name = ?", "guest").First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.logHandler.LogError(ctx, errors.New("default guest role not found"), fiber.StatusInternalServerError)
		}
		return c.logHandler.LogError(ctx, err, fiber.StatusInternalServerError)
	}

	userID := uuid.New().String()
	hashed := argon2.IDKey([]byte(req.Password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encodedHash := base64.RawStdEncoding.EncodeToString(hashed)

	user := members.User{
		ID:          userID,
		Username:    req.Username,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		DateOfBirth: dob,
		RoleID:      role.ID,
	}

	prefs := members.UserPreferences{
		UserID:   uuid.MustParse(user.ID),
		Language: "en",
		Theme:    "light",
		Timezone: "UTC",
	}

	tx := c.db.Begin()
	if tx.Error != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := c.userService.CreateUser(tx, &user, encodedHash, &prefs, nil); err != nil {
		tx.Rollback()
		return c.logHandler.LogError(ctx, err, fiber.StatusBadRequest)
	}

	if err := tx.Commit().Error; err != nil {
		return c.logHandler.LogError(ctx, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
	}

	return c.logHandler.LogSuccess(ctx, nil, "Account created successfully", true)
}
