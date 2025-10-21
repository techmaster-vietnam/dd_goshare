package rbac

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
	"gorm.io/gorm"
)

// CheckPermissionMiddleware kiểm tra quyền động cho tất cả route theo Core RBAC pattern
// Sử dụng: api.Use(rbac.CheckPermissionMiddleware())
func CheckPermissionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Log actual path và route template để debug
		log.Printf("DEBUG RBAC: c.Path() = %s, c.Route().Path = %s", c.Path(), c.Route().Path)
		route := c.Route().Path // path động (template), ví dụ: /api/rules/:ruleId/is-private
		method := c.Method()

		// DEBUG: Log middleware được gọi
		log.Printf("DEBUG RBAC: Checking permission for %s %s (route template)", method, route)

		// Normalize route key
		routeKey := method + " " + route

		// Check if route is unassigned and config allows public access
		if config.MakeUnassignedRoutePublic {
			if _, exists := routesRoles[routeKey]; !exists {
				log.Printf("DEBUG RBAC: Unassigned route, allowing access")
				return c.Next()
			}
		}

		// Get user roles from context
		userRoles := getUserRolesFromContext(c)
		log.Printf("DEBUG RBAC: User roles: %v", userRoles)

		if len(userRoles) == 0 {
			log.Printf("DEBUG RBAC: No user roles found, denying access")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Bạn chưa đăng nhập",
			})
		}

		// Check admin privilege first (admin bypasses all checks)
		if checkAdmin(userRoles) {
			log.Printf("DEBUG RBAC: Admin user, allowing access")
			return c.Next()
		}

		// Lấy rule từ DB theo path + method
		db, ok := c.Locals("db").(*gorm.DB)
		if !ok || db == nil {
			log.Printf("DEBUG RBAC: DB not found in context")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Không thể kiểm tra quyền (DB missing)",
			})
		}
		var rule struct {
			ID        int
			IsPrivate bool
		}
		log.Printf("DEBUG RBAC: Querying DB for rule with path='%s', method='%s'", route, method)
		err := db.Table("rules").Select("id, is_private").Where("path = ? AND method = ?", route, method).First(&rule).Error
		if err != nil {
			log.Printf("DEBUG RBAC: Rule not found for %s %s (path='%s', method='%s')", method, route, route, method)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Bạn không có quyền thực hiện tác vụ này",
			})
		}
		log.Printf("DEBUG RBAC: Found rule in DB: id=%d, is_private=%v", rule.ID, rule.IsPrivate)
		if !rule.IsPrivate {
			log.Printf("DEBUG RBAC: Route is public in DB, allow access")
			return c.Next()
		}
		// Nếu private, kiểm tra role như cũ
		allowedRoles := getRolesForRoute(route, method)
		if len(allowedRoles) == 0 {
			log.Printf("DEBUG RBAC: No allowed roles found for %s %s", method, route)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Bạn không có quyền thực hiện tác vụ này",
			})
		}
		convertedRoles := make(pmodel.Roles)
		for k, v := range allowedRoles {
			if v {
				convertedRoles[k] = true
			}
		}
		if checkUserRouteRoleIntersect(userRoles, convertedRoles) {
			log.Printf("DEBUG RBAC: Permission granted for %s %s", method, route)
			return c.Next()
		}
		log.Printf("DEBUG RBAC: Permission denied for %s %s", method, route)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Bạn không có quyền thực hiện tác vụ này",
		})
	}
}

// // Dummy các hàm dưới đây, bạn cần triển khai thực tế
func getUserRolesFromContext(c *fiber.Ctx) map[int]bool {
	userRoles := make(map[int]bool)
	userId, _ := c.Locals("user_id").(string)
	if userId != "" {
		db, ok := c.Locals("db").(*gorm.DB)
		if ok && db != nil {
			var userRolesList []struct{ RoleID int }
			tx := db.Table("user_roles").Select("role_id").Where("user_id = ?", userId).Find(&userRolesList)
			if tx.Error == nil {
				for _, ur := range userRolesList {
					userRoles[ur.RoleID] = true
				}
			}
		}
	}
	// Nếu chưa có thì fallback sang header (cho test hoặc trường hợp đặc biệt)
	if len(userRoles) == 0 {
		if hdr := c.Get("X-Roles"); hdr != "" {
			for _, r := range strings.Split(hdr, ",") {
				if t := strings.TrimSpace(r); t != "" {
					if id, err := strconv.Atoi(t); err == nil {
						userRoles[id] = true
					}
				}
			}
		}
	}
	return userRoles
}

// checkUserRouteRoleIntersect kiểm tra user có role phù hợp với route không
// Chỉ sử dụng cho RequireFunction - đơn giản hóa logic
func checkUserRouteRoleIntersect(userRoles map[int]bool, rolesInRoute pmodel.Roles) bool {
	// Kiểm tra xem user có bất kỳ role nào được phép truy cập route này không
	for userRole := range userRoles {
		if rolesInRoute[userRole] != nil && rolesInRoute[userRole].(bool) {
			return true // User có ít nhất 1 role được phép
		}
	}
	return false // User không có role nào được phép
}

func checkAdmin(userRoles map[int]bool) bool {
	adminRoleID := getAdminRoleID()
	return userRoles[adminRoleID]
}

// getAdminRoleID returns the admin role ID from global Roles map
func getAdminRoleID() int {
	if config.HighestRole != "" {
		return Roles[strings.ToLower(config.HighestRole)]
	}
	return Roles[DEFAULT_HIGHEST_ROLE]
}
