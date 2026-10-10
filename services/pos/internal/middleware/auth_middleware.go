package middleware

import (
	"context"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/larisai/pos-service/internal/models/entity"
)

// Protected checks for a valid JWT token
func Protected() fiber.Handler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super_secret_larisai_key_2026"
	}

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "Missing or invalid authorization header",
				},
			})
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "Invalid token",
				},
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "UNAUTHORIZED",
					"message": "Invalid token claims",
				},
			})
		}

		storeID, _ := claims["store_id"].(string)
		storeName, _ := claims["store_name"].(string)

		// Set to context for handlers to use
		c.Locals("user_id", claims["user_id"])
		c.Locals("email", claims["email"])
		c.Locals("role", claims["role"])
		c.Locals("store_id", storeID)
		c.Locals("store_name", storeName)

		// Set to Go context
		userCtx := c.UserContext()
		if userCtx == nil {
			userCtx = context.Background()
		}
		if storeID != "" {
			userCtx = context.WithValue(userCtx, "store_id", storeID)
		}
		if storeName != "" {
			userCtx = context.WithValue(userCtx, "store_name", storeName)
		}
		if userID, ok := claims["user_id"].(string); ok && userID != "" {
			userCtx = context.WithValue(userCtx, "user_id", userID)
		}
		c.SetUserContext(userCtx)

		return c.Next()
	}
}

// AdminOnly requires the user to have the admin role
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != entity.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "FORBIDDEN",
					"message": "Admin access required",
				},
			})
		}
		return c.Next()
	}
}
