package rbac

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm/clause"
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

// RegisterRulesToDB tự động insert các routes as rules vào database
func RegisterRulesToDB() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var rules []models.Rule

	// ✅ DÙNG freshRoutes thay vì routesRoles để chỉ register routes từ code hiện tại
	for _, route := range freshRoutes {
		rule := models.Rule{
			// Name sẽ được set qua API, không auto-sync từ code
			Path:       route.Path,
			Method:     route.Method,
			AccessType: route.AccessType,
			IsPrivate:  route.IsPrivate,
			Service:    config.Service,
		}
		rules = append(rules, rule)
	}

	if len(rules) == 0 {
		log.Println("No fresh routes to register as rules")
		return nil
	}

	// Sử dụng ON CONFLICT DO UPDATE để cập nhật access_type, is_private, name nếu rule đã tồn tại
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}, {Name: "method"}, {Name: "service"}},
		DoUpdates: clause.AssignmentColumns([]string{"access_type", "is_private", "name"}),
	}).Create(&rules)
	if result.Error != nil {
		return fmt.Errorf("failed to register rules: %w", result.Error)
	}

	log.Printf("Registered %d rules to database (inserted or updated)", len(rules))

	// Now need to associate roles with rules
	return associateRolesWithRules()
}

// associateRolesWithRules associates roles with rules in rule_roles table
func associateRolesWithRules() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Get all rules from database
	var dbRules []models.Rule
	if err := db.Where("service = ?", config.Service).Find(&dbRules).Error; err != nil {
		return fmt.Errorf("failed to fetch rules: %w", err)
	}

	// Create rule-role associations
	var ruleRoles []models.RuleRole

	for _, rule := range dbRules {
		routeKey := rule.Method + " " + rule.Path
		if route, exists := routesRoles[routeKey]; exists {
			for roleID, allowed := range route.Roles {
				if allowed.(bool) {
					ruleRoles = append(ruleRoles, models.RuleRole{
						RuleID: rule.ID,
						RoleID: roleID,
					})
				}
			}
		}
	}

	if len(ruleRoles) > 0 {
		// Delete existing associations for this service
		db.Where("rule_id IN (SELECT id FROM rules WHERE service = ?)", config.Service).Delete(&models.RuleRole{})

		// Insert new associations
		if err := db.Create(&ruleRoles).Error; err != nil {
			return fmt.Errorf("failed to create rule-role associations: %w", err)
		}

		log.Printf("Created %d rule-role associations", len(ruleRoles))
	}

	return nil
}

// SyncRolesWithDB sync default roles with database
func SyncRolesWithDB(defaultRoles []string) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	createdCount := 0

	for _, roleName := range defaultRoles {
		var existingRole models.Role
		result := db.Where("name = ?", strings.ToLower(roleName)).First(&existingRole)

		if result.Error != nil {
			if result.Error.Error() == "record not found" {
				// Role doesn't exist, create it
				newRole := models.Role{
					Name:        strings.ToLower(roleName),
					Description: fmt.Sprintf("Default %s role", roleName),
				}

				if err := db.Create(&newRole).Error; err != nil {
					return fmt.Errorf("failed to create role %s: %w", roleName, err)
				}

				createdCount++
				log.Printf("Created default role: %s", roleName)
			} else {
				return fmt.Errorf("error checking role %s: %w", roleName, result.Error)
			}
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

// CleanupObsoleteRules xóa các rule trong DB không còn tồn tại trong code
func CleanupObsoleteRules() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// ✅ DÙNG freshRoutes thay vì routesRoles
	current := map[string]struct{}{}
	log.Println("[RBAC CLEANUP] Fresh route keys from current code session:")
	for key := range freshRoutes {
		current[key] = struct{}{}
		log.Println("  ", key)
	}

	// Lấy toàn bộ rule trong DB cho service hiện tại
	var dbRules []struct {
		ID      int64
		Method  string
		Path    string
		Service string
	}
	if err := db.Table("rules").Select("id, method, path, service").Where("service = ?", config.Service).Find(&dbRules).Error; err != nil {
		return fmt.Errorf("failed to query rules: %w", err)
	}

	// Tìm các rule không còn trong fresh code
	var obsoleteIDs []int64
	for _, rule := range dbRules {
		key := rule.Method + "|" + rule.Path + "|" + rule.Service
		log.Println("[RBAC CLEANUP] DB rule key:", key)
		if _, ok := current[key]; !ok {
			log.Printf("[RBAC CLEANUP] Obsolete rule: id=%d method=%s path=%s service=%s", rule.ID, rule.Method, rule.Path, rule.Service)
			obsoleteIDs = append(obsoleteIDs, rule.ID)
		} else {
			log.Printf("[RBAC CLEANUP] Keep rule:    id=%d method=%s path=%s service=%s", rule.ID, rule.Method, rule.Path, rule.Service)
		}
	}

	// Xóa các rule thừa
	if len(obsoleteIDs) > 0 {
		if err := db.Table("rules").Where("id IN ?", obsoleteIDs).Delete(nil).Error; err != nil {
			return fmt.Errorf("failed to delete obsolete rules: %w", err)
		}
		log.Printf("[RBAC CLEANUP] Deleted %d obsolete rules from database", len(obsoleteIDs))
	} else {
		log.Println("[RBAC CLEANUP] No obsolete rules to delete")
	}

	return nil
}
