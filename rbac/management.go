package rbac

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

// BuildPublicRoutes tự động phát hiện public routes từ registered routes (theo Core pattern)
func BuildPublicRoutes(app *fiber.App) {
	if app == nil {
		log.Println("Warning: Fiber app is nil, cannot build public routes")
		return
	}

	routes := app.GetRoutes()
	publicCount := 0

	for _, route := range routes {
		routeKey := correctRoute(route.Method + route.Path)

		// Nếu route exists trong routesRoles và IsPrivate = false thì đây là route public
		if routeInfo, ok := routesRoles[routeKey]; ok && !routeInfo.IsPrivate {
			publicRoutes[routeKey] = true
			publicCount++
		}
	}

	log.Printf("Built %d public routes from %d total routes", publicCount, len(routes))
}

// RegisterRulesToDB tự động tạo rules từ routes đã đăng ký trong code
func RegisterRulesToDB() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Printf("DEBUG: freshRoutes count: %d", len(freshRoutes))

	var rules []models.Rule

	// ✅ DÙNG freshRoutes thay vì routesRoles để chỉ register routes từ code hiện tại
	for routeKey, route := range freshRoutes {
		log.Printf("DEBUG: Processing route: %s -> %+v", routeKey, route)
		rule := models.Rule{
			// Name sẽ được set qua API, không auto-sync từ code
			Path:       route.Path,
			Method:     route.Method,
			IsPrivate:  route.IsPrivate,
			Service:    config.Service,
			AccessType: route.AccessType, // ✅ Thêm access_type từ code
		}
		rules = append(rules, rule)
	}

	if len(rules) == 0 {
		log.Println("No fresh routes to register as rules")
		return nil
	}

	log.Printf("DEBUG: Will register %d rules to DB", len(rules))

	// Build a map of all existing rules by service
	var dbRules []models.Rule
	if err := db.Where("service = ?", config.Service).Find(&dbRules).Error; err != nil {
		return fmt.Errorf("failed to query existing rules: %w", err)
	}
	dbRuleMap := make(map[string]*models.Rule) // key: method|path
	for i := range dbRules {
		key := dbRules[i].Method + "|" + dbRules[i].Path
		dbRuleMap[key] = &dbRules[i]
	}

	for _, rule := range rules {
		key := rule.Method + "|" + rule.Path
		if existingRule, ok := dbRuleMap[key]; ok {
			// Rule exists with same method/path/service, update fields
			updates := map[string]interface{}{
				"is_private": rule.IsPrivate,
			}
			validTypes := map[int]bool{1: true, 2: true, 3: true}
			if validTypes[rule.AccessType] && rule.AccessType != existingRule.AccessType {
				updates["access_type"] = rule.AccessType
				log.Printf("DEBUG: Updated access_type for rule: %s %s (from %d to %d)", rule.Method, rule.Path, existingRule.AccessType, rule.AccessType)
			} else {
				log.Printf("DEBUG: Preserved access_type for rule: %s %s (db=%d, code=%d)", rule.Method, rule.Path, existingRule.AccessType, rule.AccessType)
			}
			if err := db.Model(existingRule).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update rule %s %s: %w", rule.Method, rule.Path, err)
			}
			continue
		}

		// If not found by method/path, try to find by service and (old path or method)
		// This is a simple heuristic: if a rule for this service exists with a different path/method, update it in place
		// (In production, you may want a more robust migration map or unique name field)
		var existingRule models.Rule
		result := db.Where("service = ? AND (path = ? OR method = ?)", rule.Service, rule.Path, rule.Method).First(&existingRule)
		if result.Error == nil {
			// Update the existing rule's path/method in place
			updates := map[string]interface{}{
				"path":       rule.Path,
				"method":     rule.Method,
				"is_private": rule.IsPrivate,
			}
			validTypes := map[int]bool{1: true, 2: true, 3: true}
			if validTypes[rule.AccessType] && rule.AccessType != existingRule.AccessType {
				updates["access_type"] = rule.AccessType
				log.Printf("DEBUG: Updated access_type for rule: %s %s (from %d to %d)", rule.Method, rule.Path, existingRule.AccessType, rule.AccessType)
			} else {
				log.Printf("DEBUG: Preserved access_type for rule: %s %s (db=%d, code=%d)", rule.Method, rule.Path, existingRule.AccessType, rule.AccessType)
			}
			if err := db.Model(&existingRule).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update rule (moved) %s %s: %w", rule.Method, rule.Path, err)
			}
			log.Printf("DEBUG: Updated rule in place (moved): id=%d new=%s %s", existingRule.ID, rule.Method, rule.Path)
			continue
		}

		// Otherwise, create new rule
		if err := db.Create(&rule).Error; err != nil {
			return fmt.Errorf("failed to create rule %s %s: %w", rule.Method, rule.Path, err)
		}
		log.Printf("DEBUG: Created new rule: %s %s", rule.Method, rule.Path)
	}

	log.Printf("DEBUG: Successfully synced %d rules to DB", len(rules))
	return nil
}


// SyncRolesWithDB sync default roles with database
func SyncRolesWithDB(defaultRoles []string) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Map role names to specific IDs to match code expectations
	roleMap := map[string]int{
		"admin":     1,
		"moderator": 2,
		"user":      3,
		"guest":     4,
	}

	createdCount := 0

	for _, roleName := range defaultRoles {
		roleNameLower := strings.ToLower(roleName)
		roleID, exists := roleMap[roleNameLower]
		if !exists {
			log.Printf("Warning: Role %s not in predefined role map, skipping", roleName)
			continue
		}

		var existingRole models.Role
		result := db.Where("id = ? OR name = ?", roleID, roleNameLower).First(&existingRole)

		if result.Error != nil {
			if result.Error.Error() == "record not found" {
				// Role doesn't exist, create it
				newRole := models.Role{
					ID:          roleID,
					Name:        roleNameLower,
					Description: fmt.Sprintf("Default %s role", roleName),
				}

				if err := db.Create(&newRole).Error; err != nil {
					return fmt.Errorf("failed to create role %s: %w", roleName, err)
				}

				createdCount++
				log.Printf("Created default role: %s (ID: %d)", roleNameLower, roleID)
			} else {
				return fmt.Errorf("error checking role %s: %w", roleName, result.Error)
			}
		} else {
			log.Printf("Role %s already exists (ID: %d)", roleNameLower, existingRole.ID)
		}
	}

	if createdCount > 0 {
		log.Printf("Created %d new default roles", createdCount)
		// Reload roles into memory
		return LoadRolesFromDB()
	}

	return nil
}

// ReloadRules reload lại các rules public, dùng khi có thay đổi về rules từ database
func ReloadRules() error {
	// Clear current rules
	routesRoles = make(map[string]Route)
	pathsRoles = make(map[string]Route)
	publicRoutes = make(map[string]bool)

	// Reload from database
	if err := LoadRulesFromDB(); err != nil {
		return fmt.Errorf("failed to reload rules: %w", err)
	}

	log.Println("Successfully reloaded RBAC rules from database")
	return nil
}

// ReloadRoles reload roles from database
func ReloadRoles() error {
	// Clear current roles
	Roles = make(map[string]int)
	roleName = make(map[int]string)

	// Reload from database
	if err := LoadRolesFromDB(); err != nil {
		return fmt.Errorf("failed to reload roles: %w", err)
	}

	log.Println("Successfully reloaded RBAC roles from database")
	return nil
}

// GetRouteInfo returns route information for debugging
func GetRouteInfo(path, method string) (Route, bool) {
	routeKey := method + " " + path
	route, exists := routesRoles[routeKey]
	return route, exists
}

// GetSystemStats returns system statistics
func GetSystemStats() map[string]interface{} {
	db := GetDB()
	stats := map[string]interface{}{
		"total_roles":         len(Roles),
		"total_routes":        len(routesRoles),
		"total_public_routes": len(publicRoutes),
		"total_paths":         len(pathsRoles),
		"service":             config.Service,
		"highest_role":        config.HighestRole,
	}

	if db != nil {
		var counts struct {
			RoleCount     int64 `json:"role_count"`
			RuleCount     int64 `json:"rule_count"`
			UserRoleCount int64 `json:"user_role_count"`
		}

		db.Model(&models.Role{}).Count(&counts.RoleCount)
		db.Model(&models.Rule{}).Where("service = ?", config.Service).Count(&counts.RuleCount)
		db.Model(&models.UserRole{}).Count(&counts.UserRoleCount)

		stats["db_roles"] = counts.RoleCount
		stats["db_rules"] = counts.RuleCount
		stats["db_user_roles"] = counts.UserRoleCount
	}

	return stats
}

// // CleanupObsoleteRules xóa các rule trong DB không còn tồn tại trong code
// func CleanupObsoleteRules() error {
// 	db := GetDB()
// 	if db == nil {
// 		return fmt.Errorf("database not initialized")
// 	}

// 	// ✅ DÙNG freshRoutes thay vì routesRoles
// 	current := map[string]struct{}{}
// 	log.Println("[RBAC CLEANUP] Fresh route keys from current code session:")
// 	for key := range freshRoutes {
// 		current[key] = struct{}{}
// 		log.Println("  ", key)
// 	}

// 	// Lấy toàn bộ rule trong DB cho service hiện tại
// 	var dbRules []struct {
// 		ID      int64
// 		Method  string
// 		Path    string
// 		Service string
// 	}
// 	if err := db.Table("rules").Select("id, method, path, service").Where("service = ?", config.Service).Find(&dbRules).Error; err != nil {
// 		return fmt.Errorf("failed to query rules: %w", err)
// 	}

// 	// Tìm các rule không còn trong fresh code
// 	var obsoleteIDs []int64
// 	for _, rule := range dbRules {
// 		// ✅ Sử dụng format key giống với freshRoutes: "method path"
// 		freshKey := rule.Method + " " + rule.Path
// 		dbKey := rule.Method + "|" + rule.Path + "|" + rule.Service
// 		log.Println("[RBAC CLEANUP] DB rule key:", dbKey)
// 		if _, ok := current[freshKey]; !ok {
// 			log.Printf("[RBAC CLEANUP] Obsolete rule: id=%d method=%s path=%s service=%s", rule.ID, rule.Method, rule.Path, rule.Service)
// 			obsoleteIDs = append(obsoleteIDs, rule.ID)
// 		} else {
// 			log.Printf("[RBAC CLEANUP] Keep rule:    id=%d method=%s path=%s service=%s", rule.ID, rule.Method, rule.Path, rule.Service)
// 		}
// 	}

// 	// Xóa các rule thừa
// 	if len(obsoleteIDs) > 0 {
// 		if err := db.Table("rules").Where("id IN ?", obsoleteIDs).Delete(nil).Error; err != nil {
// 			return fmt.Errorf("failed to delete obsolete rules: %w", err)
// 		}
// 		log.Printf("[RBAC CLEANUP] Deleted %d obsolete rules from database", len(obsoleteIDs))
// 	} else {
// 		log.Println("[RBAC CLEANUP] No obsolete rules to delete")
// 	}

// 	return nil
// }
