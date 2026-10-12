package http

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/larisai/pos-service/internal/services"
)

type UploadHandler struct {
	storageSvc services.StorageService
}

func NewUploadHandler(storageSvc services.StorageService) *UploadHandler {
	return &UploadHandler{
		storageSvc: storageSvc,
	}
}

func (h *UploadHandler) RegisterRoutes(router fiber.Router) {
	router.Post("/upload/image", h.UploadImage)
	// Static route for local image fallback
	router.Static("/images", "./uploads")
}

func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		// Fallback check field "file"
		fileHeader, err = c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "File gambar wajib diunggah dengan field 'image' atau 'file'",
			})
		}
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format file tidak didukung. Harap unggah gambar JPG, PNG, atau WebP",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membaca file gambar",
		})
	}
	defer file.Close()

	imageURL, err := h.storageSvc.UploadProductImage(c.UserContext(), fileHeader.Filename, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan gambar produk",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"url":     imageURL,
		"image_url": imageURL,
		"image_path": imageURL,
	})
}
