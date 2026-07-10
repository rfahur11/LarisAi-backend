package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/services"
)

type AnalyticsHandler struct {
	analyticsSvc services.AnalyticsService
}

func NewAnalyticsHandler(analyticsSvc services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsSvc: analyticsSvc,
	}
}

func (h *AnalyticsHandler) RegisterRoutes(router fiber.Router) {
	api := router.Group("/api/v1")
	api.Get("/analytics/summary", h.GetSummary)
}

func (h *AnalyticsHandler) GetSummary(c *fiber.Ctx) error {
	summary, err := h.analyticsSvc.GetSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Gagal mengambil ringkasan analitik",
			"detail": err.Error(),
		})
	}
	return c.JSON(summary)
}
