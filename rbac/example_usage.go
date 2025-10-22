package rbac

// File này minh họa cách sử dụng RBAC mới với Allow/Forbid
// Chỉ để demo, không phải production code

import (
	"github.com/gofiber/fiber/v2"
)

// ExampleUsage minh họa cách đăng ký route với RBAC mới
func ExampleUsage(app *fiber.App) {
	// Tạo API group
	api := app.Group("/api")

	// Ví dụ 1: Chỉ cho phép admin và moderator
	Get(api, "/admin/users", Allow(1, 2), true, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Admin users list"})
	})

	// Ví dụ 2: Cấm guest user (role ID = 3), các role khác được phép
	Post(api, "/content/create", Forbid(3), true, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Content created"})
	})

	// Ví dụ 3: Cho phép tất cả user đã đăng nhập
	Get(api, "/profile", AllowAll(), true, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "User profile"})
	})

	// Ví dụ 4: Cấm tất cả (ví dụ tính năng maintenance)
	Delete(api, "/system/reset", ForbidAll(), true, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "System reset"})
	})

	// Ví dụ 5: Route public (không cần đăng nhập)
	Get(api, "/public/info", AllowAll(), false, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Public information"})
	})
}

// InitRolesExample thiết lập roles mẫu
func InitRolesExample() {
	// Giả lập load roles từ DB
	// Trong thực tế, điều này sẽ được thực hiện trong LoadRolesFromDB()
	Roles = map[string]int{
		"admin":     1,
		"moderator": 2,
		"guest":     3,
		"user":      4,
	}
}
