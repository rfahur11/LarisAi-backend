package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/services"
)

type SubscriptionHandler struct {
	subService *services.SubscriptionService
}

func NewSubscriptionHandler(subService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subService: subService}
}

func (h *SubscriptionHandler) GetStatus(c *fiber.Ctx) error {
	storeID, ok := c.Locals("store_id").(string)
	if !ok || storeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_STORE_ID",
				"message": "Store ID missing in auth context",
			},
		})
	}

	res, err := h.subService.GetStatus(c.UserContext(), storeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "SUBSCRIPTION_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func (h *SubscriptionHandler) SyncSubscription(c *fiber.Ctx) error {
	var req dto.SubscriptionSyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "INVALID_BODY",
				"message": "Invalid request body",
			},
		})
	}

	if req.StoreID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "REQUIRED_FIELD",
				"message": "store_id is required",
			},
		})
	}

	res, err := h.subService.SyncSubscription(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "SYNC_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subscription updated successfully",
		"data":    res,
	})
}

func (h *SubscriptionHandler) RegisterProtectedRoutes(app fiber.Router) {
	app.Get("/status", h.GetStatus)
}

func (h *SubscriptionHandler) RegisterPublicRoutes(app fiber.Router) {
	app.Post("/sync", h.SyncSubscription)
}
