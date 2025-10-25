package rbac

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
	"gorm.io/gorm"
)

// CheckPermissionMiddleware kiểm tra quyền động cho tất cả route theo Core RBAC pattern
// Sử dụng: api.Use(rbac.CheckPermissionMiddleware())
func CheckPermissionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles := getUserRolesFromContext(c)

		log.Printf("DEBUG RBAC: c.Path() = %s, c.Route().Path = %s", c.Path(), c.Route().Path)
		route := c.Route().Path // path động (template), ví dụ: /api/rules/:ruleId/is-private
		method := c.Method()

		// DEBUG: Log middleware được gọi
		log.Printf("DEBUG RBAC: Checking permission for %s %s (route template)", method, route)

		// Normalize route key
		routeKey := method + " " + route

		// Kiểm tra xem route có được đăng ký với RBAC không
		registeredRoute, exists := routesRoles[routeKey]
		if exists {
			log.Printf("DEBUG RBAC: Found registered route with access type: %v, is_private: %v", registeredRoute.AccessType, registeredRoute.IsPrivate)
			// Ưu tiên kiểm tra is_private: nếu true thì bắt buộc login, không cần xét access_type
			if registeredRoute.IsPrivate {
				if len(userRoles) == 0 {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"success": false,
						"error":   "Bạn chưa đăng nhập",
					})
				}
				// Đã đăng nhập, tiếp tục kiểm tra access_type
			} else {
				// Public: ai cũng truy cập, không cần đăng nhập
				return c.Next()
			}
			// Đã đăng nhập, kiểm tra access_type (is_private và access_type hoàn toàn độc lập)
			switch registeredRoute.AccessType {
			case models.AllowAll:
				// Cho phép tất cả user đã đăng nhập
				return c.Next()
			case models.ForbidAll:
				// Cấm tất cả user đã đăng nhập
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"error":   "Tạm thời không cho phép truy cập route này",
				})
			case models.Protected:
				// Chỉ role có allowed=true trong rule_role
				db, ok := c.Locals("db").(*gorm.DB)
				if !ok || db == nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"success": false,
						"error":   "Không thể kiểm tra quyền (DB missing)",
					})
				}
				var ruleID int
				db.Table("rules").Select("id").Where("path = ? AND method = ?", registeredRoute.Path, registeredRoute.Method).Scan(&ruleID)
				for userRole := range userRoles {
					var allowed *bool
					err := db.Table("rule_roles").Select("allowed").Where("rule_id = ? AND role_id = ?", ruleID, userRole).Scan(&allowed).Error
					if err == nil && allowed != nil {
						if *allowed {
							return c.Next()
						}
						// Nếu allowed=false, explicit deny, break luôn
						return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
							"success": false,
							"error":   "Bạn không có quyền thực hiện tác vụ này (explicit deny)",
						})
					}
				}
				// Không có role nào allowed=true
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"error":   "Bạn không có quyền thực hiện tác vụ này (implicit deny)",
				})
			default:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"error":   "Bạn không có quyền thực hiện tác vụ này",
				})
			}
		}

		// Fallback về cách cũ: kiểm tra từ DB
		log.Printf("DEBUG RBAC: Route '%s' not found in routesRoles (total: %d routes)", routeKey, len(routesRoles))
		log.Printf("DEBUG RBAC: Available routes in routesRoles:")
		for key := range routesRoles {
			log.Printf("  - %s", key)
		}
		return checkDatabaseBasedAccess(c, route, method)
	}
}

// checkDatabaseBasedAccess là logic cũ kiểm tra từ database
func checkDatabaseBasedAccess(c *fiber.Ctx, route string, method string) error {
	// Check if route is unassigned and config allows public access
	routeKey := method + " " + route
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
		ID         int
		IsPrivate  bool
		AccessType int
	}
	log.Printf("DEBUG RBAC: Querying DB for rule with path='%s', method='%s'", route, method)
	err := db.Table("rules").Select("id, is_private, access_type").Where("path = ? AND method = ?", route, method).First(&rule).Error
	if err != nil {
		log.Printf("DEBUG RBAC: Rule not found for %s %s (path='%s', method='%s')", method, route, route, method)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Bạn không có quyền thực hiện tác vụ này",
		})
	}
	log.Printf("DEBUG RBAC: Found rule in DB: id=%d, is_private=%v, access_type=%d", rule.ID, rule.IsPrivate, rule.AccessType)
	if !rule.IsPrivate {
		log.Printf("DEBUG RBAC: Route is public in DB, allow access")
		return c.Next()
	}

	// Không còn allow_all/forbid_all trong v2.0

	// Không còn override kiểu cũ, chỉ dùng Public, Protected, Private

	// Không còn logic cũ, chỉ dùng Public, Protected, Private
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		"success": false,
		"error":   "Bạn không có quyền thực hiện tác vụ này",
	})
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
