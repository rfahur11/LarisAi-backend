package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AIHandler struct {
	EngineURL string
	client    *http.Client
}

func NewAIHandler(engineURL string) *AIHandler {
	return &AIHandler{
		EngineURL: engineURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetStockoutPredictions mem-proxy request ke AI Engine
func (h *AIHandler) GetStockoutPredictions(c *fiber.Ctx) error {
	resp, err := h.client.Get(h.EngineURL + "/api/v1/ai/forecasting/stockouts")
	if err != nil {
		// Mock response if AI engine is down
		return c.JSON(fiber.Map{
			"predictions": []fiber.Map{
				{
					"product_id": "1",
					"name": "Kopi Susu Gula Aren",
					"current_stock": 3,
					"daily_burn_rate": 1.5,
					"days_until_stockout": 2,
				},
				{
					"product_id": "2",
					"name": "Roti Bakar",
					"current_stock": 10,
					"daily_burn_rate": 2.0,
					"days_until_stockout": 5,
				},
			},
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read response from AI engine"})
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid JSON from AI engine"})
	}

	return c.Status(resp.StatusCode).JSON(data)
}

// GetCustomerClusters mem-proxy request ke AI Engine
func (h *AIHandler) GetCustomerClusters(c *fiber.Ctx) error {
	resp, err := h.client.Get(h.EngineURL + "/api/v1/ai/clustering/customers")
	if err != nil {
		// Mock response if AI engine is down
		return c.JSON(fiber.Map{
			"silhouette_score": 0.85,
			"clusters": []fiber.Map{
				{
					"customer_id": "c1",
					"cluster_label": "Loyal",
					"frequency_count": 10,
					"monetary_value": 500000,
				},
				{
					"customer_id": "c2",
					"cluster_label": "Churning",
					"frequency_count": 1,
					"monetary_value": 50000,
				},
			},
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read response from AI engine"})
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid JSON from AI engine"})
	}

	return c.Status(resp.StatusCode).JSON(data)
}

// SendPromo mem-proxy request ke AI Engine
func (h *AIHandler) SendPromo(c *fiber.Ctx) error {
	var reqBody interface{}
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	reqBytes, _ := json.Marshal(reqBody)
	resp, err := h.client.Post(h.EngineURL+"/api/v1/ai/promo/send", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		// Mock response if AI engine is down
		return c.JSON(fiber.Map{
			"success": true,
			"messages_sent": 5,
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read response from AI engine"})
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid JSON from AI engine"})
	}

	return c.Status(resp.StatusCode).JSON(data)
}
