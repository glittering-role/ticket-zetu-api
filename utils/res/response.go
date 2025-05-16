// utils/response/response.go
package response

import "github.com/gofiber/fiber/v2"

func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func Error(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "failed",
		"message": message,
		"data":    data,
	})
}

func Warning(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  "warning",
		"message": message,
		"data":    data,
	})
}
