package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/models/dto"
	"github.com/larisai/pos-service/internal/services"
)

type POSHandler struct {
	productSvc  services.ProductService
	checkoutSvc services.CheckoutService
}

func NewPOSHandler(productSvc services.ProductService, checkoutSvc services.CheckoutService) *POSHandler {
	return &POSHandler{
		productSvc:  productSvc,
		checkoutSvc: checkoutSvc,
	}
}

func (h *POSHandler) RegisterRoutes(router fiber.Router) {
	// Product endpoints
	router.Get("/products", h.GetProducts)
	router.Post("/products", h.CreateProduct)
	router.Delete("/products/:id", h.DeleteProduct)
	router.Get("/products/scan/:barcode", h.ScanBarcode)

	// Checkout & Transaction endpoints
	router.Post("/checkout", h.Checkout)
	router.Get("/transactions", h.GetTransactions)
}

func (h *POSHandler) GetProducts(c *fiber.Ctx) error {
	search := c.Query("search", "")
	products, err := h.productSvc.GetProducts(c.Context(), search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil daftar produk",
			"detail": err.Error(),
		})
	}
	return c.JSON(products)
}

func (h *POSHandler) CreateProduct(c *fiber.Ctx) error {
	var req dto.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request tidak valid",
		})
	}

	res, err := h.productSvc.CreateProduct(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *POSHandler) DeleteProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.productSvc.DeleteProduct(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"message": "Produk berhasil diarsipkan (soft delete)",
	})
}

func (h *POSHandler) ScanBarcode(c *fiber.Ctx) error {
	barcode := c.Params("barcode")
	product, err := h.productSvc.GetProductByBarcode(c.Context(), barcode)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(product)
}

func (h *POSHandler) Checkout(c *fiber.Ctx) error {
	var req dto.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request transaksi tidak valid",
		})
	}

	res, err := h.checkoutSvc.ProcessCheckout(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *POSHandler) GetTransactions(c *fiber.Ctx) error {
	txs, err := h.checkoutSvc.GetTransactions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil riwayat transaksi",
			"detail": err.Error(),
		})
	}
	return c.JSON(txs)
}
