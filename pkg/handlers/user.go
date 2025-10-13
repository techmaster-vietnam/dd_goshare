package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/services"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// tableName is received as a parameter, not from the path
func (h *UserHandler) GetInfoByID(c *fiber.Ctx, tableName string) error {
	id := c.Params("id")
	if tableName == "" || id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid table or id"})
	}
	acc, err := h.service.GetByUserID(c.Context(), tableName, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if acc == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(acc)
}

func (h *UserHandler) GetByUsername(c *fiber.Ctx, tableName string) error {
	username := c.Params("username")
	acc, err := h.service.GetByUsername(c.Context(), tableName, username)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if acc == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(acc)
}
