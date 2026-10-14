package http

import (
	"os"

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

func (h *SubscriptionHandler) verifyAdminKey(c *fiber.Ctx) bool {
	key := c.Get("X-Admin-Key")
	expected := os.Getenv("ADMIN_SECRET_KEY")
	if expected == "" {
		expected = "WEBOZ_LARISAI_ADMIN_SECRET_2026"
	}
	return key == expected
}

func (h *SubscriptionHandler) AdminListMerchants(c *fiber.Ctx) error {
	if !h.verifyAdminKey(c) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Invalid or missing X-Admin-Key header",
			},
		})
	}

	merchants, err := h.subService.ListAllMerchants(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "LIST_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    merchants,
	})
}

func (h *SubscriptionHandler) AdminUpdateMerchantStatus(c *fiber.Ctx) error {
	if !h.verifyAdminKey(c) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": "Invalid or missing X-Admin-Key header",
			},
		})
	}

	var req dto.AdminUpdateSubscriptionRequest
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

	res, err := h.subService.AdminUpdateSubscription(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    "UPDATE_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Store subscription updated successfully",
		"data":    res,
	})
}

func (h *SubscriptionHandler) RegisterAdminRoutes(app fiber.Router) {
	app.Get("/merchants", h.AdminListMerchants)
	app.Put("/merchant-status", h.AdminUpdateMerchantStatus)
}

