package authentication

import (
	"encoding/base64"
	"fmt"
	"ticket-zetu-api/modules/users/authentication/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

// SignUp godoc
// @Summary Sign up a new user
// @Description Creates a new user account and sends a verification email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param signup body dto.SignUpRequest true "Signup request"
// @Success 200 {object} map[string]interface{} "Account created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/signup [post]
func (ac *AuthController) SignUp(c *fiber.Ctx) error {
	var req dto.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.logHandler.LogError(c, fmt.Errorf("invalid request payload: %w", err), fiber.StatusBadRequest)
	}

	userID := uuid.New().String()
	hashed := argon2.IDKey([]byte(req.Password), []byte(userID), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLength)
	encodedHash := base64.RawStdEncoding.EncodeToString(hashed)

	user, err := ac.userService.SignUp(c.Context(), req, userID, encodedHash)
	if err != nil {
		return ac.logHandler.LogError(c, err, fiber.StatusBadRequest)
	}

	verificationCode, err := ac.emailService.GenerateAndSendVerificationCode(c, user.Email, user.Username, user.ID)
	if err != nil {
		return ac.logHandler.LogError(c, fmt.Errorf("failed to send verification code: %v", err), fiber.StatusInternalServerError)
	}

	if err := ac.userService.UpdateVerificationCode(c.Context(), user.ID, verificationCode); err != nil {
		return ac.logHandler.LogError(c, fmt.Errorf("failed to store verification code: %v", err), fiber.StatusInternalServerError)
	}

	return ac.logHandler.LogSuccess(c, fiber.Map{
		"user_id": user.ID,
		"message": "Account created successfully. Please check your email for the verification code.",
	}, "Account created successfully", true)
}
