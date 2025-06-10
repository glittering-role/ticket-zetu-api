package authentication

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

// VerifyEmailRequest defines the structure for email verification requests
type VerifyEmailRequest struct {
	Token  string `json:"token" validate:"required"`
	UserID string `json:"user_id" validate:"required,uuid"`
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verifies a user's email address using the verification code
// @Tags Authentication
// @Accept json
// @Produce json
// @Param verifyEmailRequest body VerifyEmailRequest true "Verification request"
// @Success 200 {object} map[string]interface{} "Email verified successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid or expired token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/verify-email [post]
func (ac *AuthController) VerifyEmail(c *fiber.Ctx) error {
	var req VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.logHandler.LogError(c, errors.New("invalid request payload"), fiber.StatusBadRequest)
	}

	tx := ac.db.Begin()
	if tx.Error != nil {
		return ac.logHandler.LogError(c, errors.New("failed to start transaction"), fiber.StatusInternalServerError)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			ac.logHandler.LogError(c, errors.New("panic during email verification"), fiber.StatusInternalServerError)
		}
	}()

	if err := ac.userService.VerifyEmailCode(tx, req.UserID, req.Token); err != nil {
		tx.Rollback()
		return ac.logHandler.LogError(c, err, fiber.StatusUnauthorized)
	}

	if err := tx.Commit().Error; err != nil {
		return ac.logHandler.LogError(c, errors.New("failed to commit transaction"), fiber.StatusInternalServerError)
	}

	return ac.logHandler.LogSuccess(c, fiber.Map{
		"message": "Email verified successfully",
	}, "Email verified successfully", false)
}
