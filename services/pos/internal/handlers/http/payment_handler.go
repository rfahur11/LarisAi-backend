package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/repositories"
)

type PaymentHandler struct {
	txRepo repositories.TransactionRepository
}

func NewPaymentHandler(txRepo repositories.TransactionRepository) *PaymentHandler {
	return &PaymentHandler{
		txRepo: txRepo,
	}
}

func (h *PaymentHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/webhook", h.HandleWebhook)
}

func (h *PaymentHandler) HandleWebhook(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	orderID, ok := payload["order_id"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing order_id"})
	}

	transactionStatus, ok := payload["transaction_status"].(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing transaction_status"})
	}

	var status string
	switch transactionStatus {
	case "capture", "settlement":
		status = "paid"
	case "pending":
		status = "pending"
	case "deny", "cancel", "expire", "failure":
		status = "failed"
	default:
		status = "unknown"
	}

	log.Printf("Received Midtrans webhook: OrderID=%s, Status=%s", orderID, status)

	err := h.txRepo.UpdateStatus(c.Context(), orderID, status)
	if err != nil {
		log.Printf("Failed to update transaction status: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"status": "success"})
}
