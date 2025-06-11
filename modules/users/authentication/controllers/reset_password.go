package authentication

import (
	"errors"
	"ticket-zetu-api/modules/users/authentication/dto"

	"github.com/gofiber/fiber/v2"
)

// RequestPasswordReset godoc
// @Summary Request a password reset
// @Description Sends a password reset link to the user's email based on username or email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Username or email for password reset"
// @Success 200 {object} map[string]interface{} "Password reset email sent"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 423 {object} map[string]interface{} "Account locked"
// @Failure 429 {object} map[string]interface{} "Too many requests"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/reset-password-request [post]
func (ac *AuthController) RequestPasswordReset(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.logHandler.LogError(c, errors.New("invalid request payload"), fiber.StatusBadRequest)
	}

	err := ac.userService.RequestPasswordReset(c.Context(), c, req.UsernameOrEmail)
	if err != nil {
		return ac.logHandler.LogError(c, err, getStatusCodeFromError(err))
	}

	return ac.logHandler.LogSuccess(c, nil, "Password reset email sent", true)
}

// SetNewPassword godoc
// @Summary Set a new password
// @Description Updates the user's password using a valid reset token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.SetNewPasswordRequest true "Reset token and new password"
// @Success 200 {object} map[string]interface{} "Password updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid or expired reset token"
// @Failure 423 {object} map[string]interface{} "Account locked"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/reset-password [post]
func (ac *AuthController) SetNewPassword(c *fiber.Ctx) error {
	var req dto.SetNewPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.logHandler.LogError(c, errors.New("invalid request payload"), fiber.StatusBadRequest)
	}

	err := ac.userService.SetNewPassword(c.Context(), c, req.ResetToken, req.NewPassword)
	if err != nil {
		return ac.logHandler.LogError(c, err, getStatusCodeFromError(err))
	}

	return ac.logHandler.LogSuccess(c, nil, "Password updated successfully", true)
}
