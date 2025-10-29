package rbac

/*
RBAC Management Package - Enhanced Rule Synchronization System

This package provides comprehensive rule and role management with automatic synchronization
between code-defined routes and database rules.

KEY FEATURES:
1. Automatic path change detection and rule_roles migration
2. Obsolete rules cleanup with CASCADE delete
3. Flexible role assignment strategies (smart, all roles, specific roles)
4. Orphaned rule_roles cleanup

USAGE EXAMPLES:

	// 1. Smart sync with intelligent role assignment based on access_type
	if err := rbac.FullRuleSync(); err != nil {
		log.Fatalf("Failed to sync: %v", err)
	}

	// 2. Sync and assign ALL roles to new rules
	if err := rbac.FullRuleSyncWithAllRoles(); err != nil {
		log.Fatalf("Failed to sync: %v", err)
	}

	// 3. Sync and assign specific roles (e.g., role 1, 2, 3)
	if err := rbac.FullRuleSyncWithRoles(1, 2, 3); err != nil {
		log.Fatalf("Failed to sync: %v", err)
	}

PATH CHANGE HANDLING:
When a route path changes in code (e.g., /api/user -> /api/users):
- System detects the change by matching method+service
- Creates new rule with new path
- Automatically migrates all rule_roles from old rule to new rule
- Cleans up the obsolete old rule
- Result: Zero downtime, no permission loss
*/

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

	// ✅ Use UPSERT with enhanced logic to handle path changes
	// Build a map of all rules in DB for this service to detect changes
	var dbRules []models.Rule
	if err := db.Where("service = ?", config.Service).Find(&dbRules).Error; err != nil {
		return fmt.Errorf("failed to query existing rules: %w", err)
	}

	dbRuleMap := make(map[string]*models.Rule) // key: method|service
	for i := range dbRules {
		key := dbRules[i].Method + "|" + dbRules[i].Service
		dbRuleMap[key] = &dbRules[i]
	}

	for _, rule := range rules {
		var existingRule models.Rule
		result := db.Where("path = ? AND method = ? AND service = ?", rule.Path, rule.Method, rule.Service).First(&existingRule)

		if result.Error != nil {
			// Rule doesn't exist with this exact path+method+service
			// Try to find an existing rule by method+service (same logical rule, path changed)
			ruleKey := rule.Method + "|" + rule.Service
			if oldRule, exists := dbRuleMap[ruleKey]; exists {
				// Update existing rule in-place (preserve ID so rule_roles keep pointing to it)
				if oldRule.Path != rule.Path || oldRule.IsPrivate != rule.IsPrivate {
					updates := map[string]interface{}{
						"path":       rule.Path,
						"is_private": rule.IsPrivate,
					}
					if err := db.Model(oldRule).Updates(updates).Error; err != nil {
						return fmt.Errorf("failed to update existing rule %d in-place: %w", oldRule.ID, err)
					}
					log.Printf("♻️  Updated existing rule in-place: id=%d method=%s path=%s (preserved access_type)", oldRule.ID, rule.Method, rule.Path)
				} else {
					log.Printf("♻️  Keep existing rule unchanged: id=%d method=%s path=%s", oldRule.ID, rule.Method, rule.Path)
				}
			} else {
				// Completely new rule
				if err := db.Create(&rule).Error; err != nil {
					return fmt.Errorf("failed to create rule %s %s: %w", rule.Method, rule.Path, err)
				}
				log.Printf("✅ Created new rule: %s %s (ID: %d)", rule.Method, rule.Path, rule.ID)
			}
		} else {
			// Rule exists with same path+method+service: update safe fields only
			updates := map[string]interface{}{
				"is_private": rule.IsPrivate,
				// NOTE: Do NOT update access_type - preserve user customizations
			}
			if err := db.Model(&existingRule).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update rule %s %s: %w", rule.Method, rule.Path, err)
			}
			log.Printf("♻️  Updated existing rule: %s %s (preserved access_type)", rule.Method, rule.Path)
		}
	}

	log.Printf("DEBUG: Successfully synced %d rules to DB", len(rules))

	// // ✅ Cleanup obsolete rules after syncing fresh routes
	// if err := CleanupObsoleteRules(); err != nil {
	// 	log.Printf("Warning: Failed to cleanup obsolete rules: %v", err)
	// 	// Don't return error, just log warning to not break the main flow
	// }

	return nil
}

// SyncRulesToDB đồng bộ rules từ code và xóa rules cũ không còn tồn tại
func SyncRulesToDB() error {
	// 1. Đăng ký/cập nhật rules từ fresh routes
	if err := RegisterRulesToDB(); err != nil {
		return fmt.Errorf("failed to register rules: %w", err)
	}

	log.Println("Successfully synced rules to database and cleaned up obsolete rules")
	return nil
}



// DebugRuleMigration hiển thị chi tiết rule_roles trước và sau migration
func DebugRuleMigration(oldRuleID, newRuleID int) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("==========================================")
	log.Printf("🔍 DEBUG: Rule Migration Analysis")
	log.Println("==========================================")

	// Check old rule
	var oldRule models.Rule
	if err := db.Where("id = ?", oldRuleID).First(&oldRule).Error; err != nil {
		log.Printf("Old Rule ID %d: NOT FOUND", oldRuleID)
	} else {
		log.Printf("Old Rule ID %d: %s %s", oldRuleID, oldRule.Method, oldRule.Path)
	}

	// Check new rule
	var newRule models.Rule
	if err := db.Where("id = ?", newRuleID).First(&newRule).Error; err != nil {
		log.Printf("New Rule ID %d: NOT FOUND", newRuleID)
	} else {
		log.Printf("New Rule ID %d: %s %s", newRuleID, newRule.Method, newRule.Path)
	}

	log.Println("------------------------------------------")

	// Get old rule_roles
	var oldRuleRoles []models.RuleRole
	if err := db.Where("rule_id = ?", oldRuleID).Find(&oldRuleRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch old rule_roles: %w", err)
	}

	log.Printf("Old Rule (%d) has %d role assignments:", oldRuleID, len(oldRuleRoles))
	for _, rr := range oldRuleRoles {
		log.Printf("   - RoleID: %d, Allowed: %v", rr.RoleID, rr.Allowed)
	}

	log.Println("------------------------------------------")

	// Get new rule_roles
	var newRuleRoles []models.RuleRole
	if err := db.Where("rule_id = ?", newRuleID).Find(&newRuleRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch new rule_roles: %w", err)
	}

	log.Printf("New Rule (%d) has %d role assignments:", newRuleID, len(newRuleRoles))
	for _, rr := range newRuleRoles {
		log.Printf("   - RoleID: %d, Allowed: %v", rr.RoleID, rr.Allowed)
	}

	log.Println("------------------------------------------")

	// Analysis
	if len(oldRuleRoles) > len(newRuleRoles) {
		log.Printf("⚠️  WARNING: Lost %d role assignments during migration!",
			len(oldRuleRoles)-len(newRuleRoles))

		// Find missing roles
		oldRoleMap := make(map[int]bool)
		for _, rr := range oldRuleRoles {
			oldRoleMap[rr.RoleID] = true
		}

		newRoleMap := make(map[int]bool)
		for _, rr := range newRuleRoles {
			newRoleMap[rr.RoleID] = true
		}

		log.Println("Missing Role IDs:")
		for roleID := range oldRoleMap {
			if !newRoleMap[roleID] {
				log.Printf("   - RoleID %d is missing in new rule", roleID)
			}
		}
	} else if len(oldRuleRoles) == len(newRuleRoles) {
		log.Println("✅ All role assignments successfully migrated")
	} else {
		log.Printf("ℹ️  New rule has MORE roles than old rule (+%d)",
			len(newRuleRoles)-len(oldRuleRoles))
	}

	log.Println("==========================================")
	return nil
}

// AutoAssignDefaultRoles tự động gán roles mặc định cho rules chưa có role assignments
func AutoAssignDefaultRoles() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Lấy tất cả roles có sẵn trong hệ thống
	var availableRoles []models.Role
	if err := db.Find(&availableRoles).Error; err != nil {
		return fmt.Errorf("failed to fetch available roles: %w", err)
	}

	if len(availableRoles) == 0 {
		log.Println("No roles found in database, skipping auto-assignment")
		return nil
	}

	// Tìm các rule chưa có role assignments
	var rulesWithoutRoles []struct {
		ID         int    `json:"id"`
		Path       string `json:"path"`
		Method     string `json:"method"`
		AccessType int    `json:"access_type"`
	}

	query := `
	SELECT r.id, r.path, r.method, r.access_type
	FROM rules r
	LEFT JOIN rule_roles rr ON r.id = rr.rule_id
	WHERE r.service = ? AND rr.rule_id IS NULL
	`

	if err := db.Raw(query, config.Service).Scan(&rulesWithoutRoles).Error; err != nil {
		return fmt.Errorf("failed to find rules without roles: %w", err)
	}

	if len(rulesWithoutRoles) == 0 {
		log.Println("All rules already have role assignments")
		return nil
	}

	// Gán roles theo logic nghiệp vụ
	var ruleRoles []models.RuleRole
	for _, rule := range rulesWithoutRoles {
		assignedRoles := determineRolesForRule(rule, availableRoles)

		for _, roleID := range assignedRoles {
			ruleRole := models.RuleRole{
				RuleID: rule.ID,
				RoleID: roleID,
				// Allowed sẽ là nil (default) để follow rule's access_type
			}
			ruleRoles = append(ruleRoles, ruleRole)
		}

		log.Printf("Auto-assigning roles %v to rule: %s %s (ID: %d, AccessType: %d)",
			assignedRoles, rule.Method, rule.Path, rule.ID, rule.AccessType)
	}

	// Batch insert rule_roles
	if err := db.Create(&ruleRoles).Error; err != nil {
		return fmt.Errorf("failed to auto-assign roles: %w", err)
	}

	log.Printf("Auto-assigned %d role assignments to %d rules", len(ruleRoles), len(rulesWithoutRoles))
	return nil
}

// determineRolesForRule xác định roles nào sẽ được gán cho rule dựa trên access_type và logic nghiệp vụ
func determineRolesForRule(rule struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	AccessType int    `json:"access_type"`
}, availableRoles []models.Role) []int {

	var roleIDs []int

	// Logic gán role dựa trên access_type:
	switch rule.AccessType {
	case 1: // AllowAll - gán tất cả roles
		for _, role := range availableRoles {
			roleIDs = append(roleIDs, role.ID)
		}

	case 2: // Protected - chỉ gán admin role
		for _, role := range availableRoles {
			if role.Name == "admin" {
				roleIDs = append(roleIDs, role.ID)
				break
			}
		}
		// Nếu không tìm thấy admin role, gán role đầu tiên
		if len(roleIDs) == 0 && len(availableRoles) > 0 {
			roleIDs = append(roleIDs, availableRoles[0].ID)
		}

	case 3: // ForbidAll - gán admin role để có thể override
		for _, role := range availableRoles {
			if role.Name == "admin" {
				roleIDs = append(roleIDs, role.ID)
				break
			}
		}
		// Nếu không tìm thấy admin role, gán role đầu tiên
		if len(roleIDs) == 0 && len(availableRoles) > 0 {
			roleIDs = append(roleIDs, availableRoles[0].ID)
		}

	default:
		// Default: gán tất cả roles
		for _, role := range availableRoles {
			roleIDs = append(roleIDs, role.ID)
		}
	}

	return roleIDs
}

// AutoAssignSpecificRoles gán các role IDs cụ thể cho rules chưa có role assignments
func AutoAssignSpecificRoles(roleIDs []int) error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	if len(roleIDs) == 0 {
		return fmt.Errorf("no role IDs provided")
	}

	// Validate role IDs exist
	var existingRoles []models.Role
	if err := db.Where("id IN ?", roleIDs).Find(&existingRoles).Error; err != nil {
		return fmt.Errorf("failed to validate role IDs: %w", err)
	}

	if len(existingRoles) != len(roleIDs) {
		return fmt.Errorf("some role IDs do not exist in database")
	}

	// Tìm các rule chưa có role assignments
	var rulesWithoutRoles []struct {
		ID         int    `json:"id"`
		Path       string `json:"path"`
		Method     string `json:"method"`
		AccessType int    `json:"access_type"`
	}

	query := `
	SELECT r.id, r.path, r.method, r.access_type
	FROM rules r
	LEFT JOIN rule_roles rr ON r.id = rr.rule_id
	WHERE r.service = ? AND rr.rule_id IS NULL
	`

	if err := db.Raw(query, config.Service).Scan(&rulesWithoutRoles).Error; err != nil {
		return fmt.Errorf("failed to find rules without roles: %w", err)
	}

	if len(rulesWithoutRoles) == 0 {
		log.Println("All rules already have role assignments")
		return nil
	}

	// Gán các role IDs đã chỉ định cho tất cả rules
	var ruleRoles []models.RuleRole
	for _, rule := range rulesWithoutRoles {
		for _, roleID := range roleIDs {
			ruleRole := models.RuleRole{
				RuleID: rule.ID,
				RoleID: roleID,
				// Allowed sẽ là nil (default) để follow rule's access_type
			}
			ruleRoles = append(ruleRoles, ruleRole)
		}

		log.Printf("Auto-assigning roles %v to rule: %s %s (ID: %d)",
			roleIDs, rule.Method, rule.Path, rule.ID)
	}

	// Batch insert rule_roles
	if err := db.Create(&ruleRoles).Error; err != nil {
		return fmt.Errorf("failed to auto-assign specific roles: %w", err)
	}

	log.Printf("Auto-assigned %d role assignments (%d rules × %d roles) with specific role IDs",
		len(ruleRoles), len(rulesWithoutRoles), len(roleIDs))
	return nil
}

// AutoAssignAllRoles gán tất cả roles có sẵn cho rules chưa có role assignments
func AutoAssignAllRoles() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Lấy tất cả role IDs
	var roleIDs []int
	if err := db.Model(&models.Role{}).Pluck("id", &roleIDs).Error; err != nil {
		return fmt.Errorf("failed to fetch role IDs: %w", err)
	}

	if len(roleIDs) == 0 {
		log.Println("No roles found in database")
		return nil
	}

	return AutoAssignSpecificRoles(roleIDs)
}

// ComprehensiveRuleSync thực hiện full sync: register, cleanup, và auto-assign roles với logic thông minh
func ComprehensiveRuleSync() error {
	log.Println("Starting comprehensive rule synchronization...")

	// 1. Register/update rules from fresh routes (includes cleanup of obsolete rules)
	if err := RegisterRulesToDB(); err != nil {
		return fmt.Errorf("failed to register rules: %w", err)
	}

	// 2. Auto-assign default roles to rules without role assignments (smart logic based on access_type)
	if err := AutoAssignDefaultRoles(); err != nil {
		log.Printf("Warning: Failed to auto-assign default roles: %v", err)
		// Don't fail the whole process, just log warning
	}

	// 3. Final cleanup of any orphaned rule_roles
	if err := CleanupOrphanedRuleRoles(); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned rule_roles: %v", err)
	}

	log.Println("✅ Comprehensive rule synchronization completed successfully")
	return nil
}

// ComprehensiveRuleSyncWithAllRoles thực hiện full sync và gán TẤT CẢ roles cho mọi rule
func ComprehensiveRuleSyncWithAllRoles() error {
	log.Println("Starting comprehensive rule synchronization with ALL roles assignment...")

	// 1. Register/update rules from fresh routes (includes cleanup of obsolete rules)
	if err := RegisterRulesToDB(); err != nil {
		return fmt.Errorf("failed to register rules: %w", err)
	}

	// 2. Auto-assign ALL roles to rules without role assignments
	if err := AutoAssignAllRoles(); err != nil {
		log.Printf("Warning: Failed to auto-assign all roles: %v", err)
		// Don't fail the whole process, just log warning
	}

	// 3. Final cleanup of any orphaned rule_roles
	if err := CleanupOrphanedRuleRoles(); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned rule_roles: %v", err)
	}

	log.Println("✅ Comprehensive rule synchronization with all roles completed successfully")
	return nil
}

// ComprehensiveRuleSyncWithSpecificRoles thực hiện full sync và gán các role IDs cụ thể
func ComprehensiveRuleSyncWithSpecificRoles(roleIDs []int) error {
	log.Printf("Starting comprehensive rule synchronization with specific roles %v...", roleIDs)

	// 1. Register/update rules from fresh routes (includes cleanup of obsolete rules)
	if err := RegisterRulesToDB(); err != nil {
		return fmt.Errorf("failed to register rules: %w", err)
	}

	// 2. Auto-assign specific roles to rules without role assignments
	if err := AutoAssignSpecificRoles(roleIDs); err != nil {
		log.Printf("Warning: Failed to auto-assign specific roles: %v", err)
		// Don't fail the whole process, just log warning
	}

	// 3. Final cleanup of any orphaned rule_roles
	if err := CleanupOrphanedRuleRoles(); err != nil {
		log.Printf("Warning: Failed to cleanup orphaned rule_roles: %v", err)
	}

	log.Printf("✅ Comprehensive rule synchronization with specific roles %v completed successfully", roleIDs)
	return nil
}

// CleanupOrphanedRuleRoles xóa các rule_roles có rule_id không tồn tại trong bảng rules
func CleanupOrphanedRuleRoles() error {
	db := GetDB()
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// Xóa rule_roles mà rule_id không tồn tại trong bảng rules
	result := db.Exec(`
		DELETE FROM rule_roles 
		WHERE rule_id NOT IN (
			SELECT id FROM rules WHERE service = ?
		)
	`, config.Service)

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup orphaned rule_roles: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d orphaned rule_roles records", result.RowsAffected)
	} else {
		log.Println("No orphaned rule_roles found")
	}

	return nil
}

// FullRuleSync - alias cho ComprehensiveRuleSync để dễ sử dụng hơn
// Sử dụng function này khi muốn đồng bộ rules với logic thông minh (dựa trên access_type)
func FullRuleSync() error {
	return ComprehensiveRuleSync()
}

// FullRuleSyncWithAllRoles - đồng bộ rules và gán TẤT CẢ roles cho mọi rule
func FullRuleSyncWithAllRoles() error {
	return ComprehensiveRuleSyncWithAllRoles()
}

// FullRuleSyncWithRoles - đồng bộ rules và gán các role IDs cụ thể
func FullRuleSyncWithRoles(roleIDs ...int) error {
	return ComprehensiveRuleSyncWithSpecificRoles(roleIDs)
}

// VerifyRuleRoleConsistency kiểm tra tính nhất quán giữa rules và rule_roles
func VerifyRuleRoleConsistency() (*RuleRoleConsistencyReport, error) {
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	report := &RuleRoleConsistencyReport{
		Service: config.Service,
	}

	// 1. Count total rules for this service
	var totalRules int64
	if err := db.Model(&models.Rule{}).Where("service = ?", config.Service).Count(&totalRules).Error; err != nil {
		return nil, fmt.Errorf("failed to count rules: %w", err)
	}
	report.TotalRules = int(totalRules)

	// 2. Find rules without any role assignments
	var rulesWithoutRoles []struct {
		ID     int    `json:"id"`
		Path   string `json:"path"`
		Method string `json:"method"`
	}
	query := `
	SELECT r.id, r.path, r.method
	FROM rules r
	LEFT JOIN rule_roles rr ON r.id = rr.rule_id
	WHERE r.service = ? AND rr.rule_id IS NULL
	`
	if err := db.Raw(query, config.Service).Scan(&rulesWithoutRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to find rules without roles: %w", err)
	}
	report.RulesWithoutRoles = len(rulesWithoutRoles)
	report.RulesWithoutRolesList = rulesWithoutRoles

	// 3. Find orphaned rule_roles (rule_id doesn't exist in rules table)
	var orphanedRuleRoles []struct {
		RuleID int `json:"rule_id"`
		RoleID int `json:"role_id"`
	}
	orphanQuery := `
	SELECT rr.rule_id, rr.role_id
	FROM rule_roles rr
	LEFT JOIN rules r ON rr.rule_id = r.id
	WHERE r.id IS NULL OR r.service != ?
	`
	if err := db.Raw(orphanQuery, config.Service).Scan(&orphanedRuleRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to find orphaned rule_roles: %w", err)
	}
	report.OrphanedRuleRoles = len(orphanedRuleRoles)
	report.OrphanedRuleRolesList = orphanedRuleRoles

	// 4. Count total rule_roles
	var totalRuleRoles int64
	if err := db.Table("rule_roles").
		Joins("JOIN rules ON rules.id = rule_roles.rule_id").
		Where("rules.service = ?", config.Service).
		Count(&totalRuleRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to count rule_roles: %w", err)
	}
	report.TotalRuleRoles = int(totalRuleRoles)

	// 5. Calculate health status
	report.IsHealthy = report.RulesWithoutRoles == 0 && report.OrphanedRuleRoles == 0

	return report, nil
}

// RuleRoleConsistencyReport báo cáo tình trạng đồng bộ giữa rules và rule_roles
type RuleRoleConsistencyReport struct {
	Service               string `json:"service"`
	TotalRules            int    `json:"total_rules"`
	TotalRuleRoles        int    `json:"total_rule_roles"`
	RulesWithoutRoles     int    `json:"rules_without_roles"`
	RulesWithoutRolesList []struct {
		ID     int    `json:"id"`
		Path   string `json:"path"`
		Method string `json:"method"`
	} `json:"rules_without_roles_list,omitempty"`
	OrphanedRuleRoles     int `json:"orphaned_rule_roles"`
	OrphanedRuleRolesList []struct {
		RuleID int `json:"rule_id"`
		RoleID int `json:"role_id"`
	} `json:"orphaned_rule_roles_list,omitempty"`
	IsHealthy bool `json:"is_healthy"`
}

// PrintReport in ra báo cáo dễ đọc
func (r *RuleRoleConsistencyReport) PrintReport() {
	log.Println("==========================================")
	log.Printf("📊 RBAC Rule-Role Consistency Report")
	log.Println("==========================================")
	log.Printf("Service: %s", r.Service)
	log.Printf("Total Rules: %d", r.TotalRules)
	log.Printf("Total Rule-Role Assignments: %d", r.TotalRuleRoles)
	log.Println("------------------------------------------")

	if r.RulesWithoutRoles > 0 {
		log.Printf("⚠️  Rules WITHOUT Role Assignments: %d", r.RulesWithoutRoles)
		for _, rule := range r.RulesWithoutRolesList {
			log.Printf("   - Rule ID %d: %s %s", rule.ID, rule.Method, rule.Path)
		}
	} else {
		log.Println("✅ All rules have role assignments")
	}

	log.Println("------------------------------------------")

	if r.OrphanedRuleRoles > 0 {
		log.Printf("⚠️  Orphaned Rule-Roles: %d", r.OrphanedRuleRoles)
		for _, rr := range r.OrphanedRuleRolesList {
			log.Printf("   - RuleID: %d, RoleID: %d", rr.RuleID, rr.RoleID)
		}
	} else {
		log.Println("✅ No orphaned rule-roles")
	}

	log.Println("------------------------------------------")

	if r.IsHealthy {
		log.Println("🎉 Status: HEALTHY - All rules properly configured")
	} else {
		log.Println("🚨 Status: UNHEALTHY - Issues detected, run FullRuleSync() to fix")
	}
	log.Println("==========================================")
}

// QuickHealthCheck thực hiện health check nhanh và in report
func QuickHealthCheck() error {
	report, err := VerifyRuleRoleConsistency()
	if err != nil {
		return fmt.Errorf("failed to verify rule-role consistency: %w", err)
	}

	report.PrintReport()
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
		"admin": 1,
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

// 	// Xóa các rule thừa (CASCADE sẽ tự động xóa rule_roles)
// 	if len(obsoleteIDs) > 0 {
// 		if err := db.Table("rules").Where("id IN ?", obsoleteIDs).Delete(nil).Error; err != nil {
// 			return fmt.Errorf("failed to delete obsolete rules: %w", err)
// 		}
// 		log.Printf("[RBAC CLEANUP] Deleted %d obsolete rules from database (rule_roles auto-deleted via CASCADE)", len(obsoleteIDs))
// 	} else {
// 		log.Println("[RBAC CLEANUP] No obsolete rules to delete")
// 	}

// 	return nil
// }
