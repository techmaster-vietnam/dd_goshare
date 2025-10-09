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
		route := c.Route().Path
		method := c.Method()

		// DEBUG: Log middleware được gọi
		log.Printf("DEBUG RBAC: Checking permission for %s %s", method, route)

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

		// Get route roles and check permission using Core RBAC logic
		if routeInfo, exists := routesRoles[routeKey]; exists {
			log.Printf("DEBUG RBAC: Route info found: %+v", routeInfo)

			// If Roles snapshot is empty, dynamically fetch allowed roles by business function (access_type)
			if len(routeInfo.Roles) == 0 && strings.TrimSpace(routeInfo.AccessType) != "" {
				log.Printf("DEBUG RBAC: Empty Roles snapshot for route, loading allowed roles for access_type='%s' from DB", routeInfo.AccessType)
				allowed := getRolesForFunction(routeInfo.AccessType)
				if len(allowed) > 0 {
					// Build Roles map and cache back to route for next time
					newRoles := make(pmodel.Roles)
					for roleID, ok := range allowed {
						if ok {
							newRoles[roleID] = true
						}
					}
					routeInfo.Roles = newRoles
					routesRoles[routeKey] = routeInfo
					log.Printf("DEBUG RBAC: Loaded and cached %d allowed roles for %s", len(newRoles), routeKey)
				} else {
					log.Printf("DEBUG RBAC: No allowed roles found in DB for access_type='%s'", routeInfo.AccessType)
				}
			}

			if checkUserRouteRoleIntersect(userRoles, routeInfo.Roles) {
				log.Printf("DEBUG RBAC: Permission granted")
				return c.Next()
			}
			log.Printf("DEBUG RBAC: Permission denied")
		} else {
			log.Printf("DEBUG RBAC: No route info found for %s", routeKey)
			// If route not found and MakeUnassignedRoutePublic is false, deny access
			if !config.MakeUnassignedRoutePublic {
				log.Printf("DEBUG RBAC: Denying access to unassigned route")
			} else {
				log.Printf("DEBUG RBAC: Allowing access to unassigned route")
				return c.Next()
			}
		}

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
