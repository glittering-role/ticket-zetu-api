package authentication

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"ticket-zetu-api/modules/users/authentication/dto"
)

// SignIn godoc
// @Summary Authenticate a user
// @Description Logs in a user and returns session tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 403 {object} map[string]interface{} "Email not verified"
// @Failure 423 {object} map[string]interface{} "Account locked"
// @Failure 429 {object} map[string]interface{} "Too many resend attempts"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/signin [post]
func (ac *AuthController) SignIn(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.logHandler.LogError(c, fmt.Errorf("invalid request payload: %w", err), fiber.StatusBadRequest)
	}

	// Pass raw password to Authenticate
	_, session, err := ac.userService.Authenticate(c.Context(), c, req.UsernameOrEmail, req.Password, req.RememberMe, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return ac.logHandler.LogError(c, err, getStatusCodeFromError(err))
	}

	// Set session cookies
	isProd := true
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    session.SessionToken,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "Strict",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    session.RefreshToken,
		Expires:  session.RefreshExpiry,
		HTTPOnly: true,
		Secure:   isProd,
		SameSite: "Strict",
	})

	return ac.logHandler.LogSuccess(c, nil, "Login successful", true)
}

// getStatusCodeFromError maps error messages to HTTP status codes
func getStatusCodeFromError(err error) int {
	switch {
	case err.Error() == "invalid credentials":
		return fiber.StatusUnauthorized
	case err.Error() == "email not verified, verification email resent":
		return fiber.StatusForbidden
	case err.Error() == "account temporarily locked" || err.Error() == "account locked due to too many failed attempts":
		return fiber.StatusLocked
	case err.Error() == "email verification resend limit exceeded, try again later":
		return fiber.StatusForbidden
	case err.Error() == "please wait before requesting another verification email":
		return fiber.StatusTooManyRequests
	default:
		return fiber.StatusInternalServerError
	}
}
