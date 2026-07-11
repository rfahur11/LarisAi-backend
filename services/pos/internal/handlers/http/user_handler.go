package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	users, err := h.userService.GetAllUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch users",
		})
	}
	
	// Create response DTO (strip passwords)
	var response []map[string]interface{}
	for _, u := range users {
		response = append(response, map[string]interface{}{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"role":       u.Role,
			"created_at": u.CreatedAt,
		})
	}

	// In case the list is empty
	if response == nil {
		response = make([]map[string]interface{}, 0)
	}

	return c.JSON(response)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Name, email, and password are required",
		})
	}

	err := h.userService.CreateUser(c.Context(), req)
	if err != nil {
		if err.Error() == "email already exists" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message": "Email already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "ID parameter is required",
		})
	}

	err := h.userService.DeleteUser(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
