package database

import (
	"fmt"

	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)
// hàm này sẽ tạo role admin và gán tất cả các rule hiện có cho role này, sau đó gán role admin cho user đầu tiên tạo tài khoản
func CreateAdminRoleAndAssign(db *gorm.DB, userID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Create admin role if not exists
		var adminRole models.Role
		if err := tx.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				adminRole = models.Role{
					Name:        "admin",
					Description: "Administrator role with full system access",
				}
				if err := tx.Create(&adminRole).Error; err != nil {
					return fmt.Errorf("failed to create admin role: %v", err)
				}
			} else {
				return fmt.Errorf("error checking admin role: %v", err)
			}
		}

		// 2. Get all available rules
		var allRules []models.Rule
		if err := tx.Find(&allRules).Error; err != nil {
			return fmt.Errorf("failed to get rules: %v", err)
		}

		// 3. Create role-rule assignments for all rules
		for _, rule := range allRules {
			var existingAssignment models.RuleRole
			if err := tx.Where("role_id = ? AND rule_id = ?", adminRole.ID, rule.ID).First(&existingAssignment).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					assignment := models.RuleRole{
						RoleID: adminRole.ID,
						RuleID: rule.ID,
					}
					if err := tx.Create(&assignment).Error; err != nil {
						return fmt.Errorf("failed to assign rule to admin role")
					}
				} else {
					return fmt.Errorf("error checking role-rule assignment: %v", err)
				}
			}
		}

		// 4. Assign admin role to user
		var existingUserRole models.UserRole
		if err := tx.Where("user_id = ? AND role_id = ?", userID, adminRole.ID).First(&existingUserRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				userRole := models.UserRole{
					UserID: userID,
					RoleID: adminRole.ID,
				}
				if err := tx.Create(&userRole).Error; err != nil {
					return fmt.Errorf("failed to assign admin role to user: %v", err)
				}
			} else {
				return fmt.Errorf("error checking user role assignment: %v", err)
			}
		}

		return nil
	})
}

// GetUserPermissions returns all permissions for a user based on their roles
func GetUserPermissions(db *gorm.DB, userID string) ([]string, error) {
	var permissions []string

	// Get user roles
	var userRoles []models.UserRole
	if err := db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %v", err)
	}

	// Get role IDs
	var roleIDs []int
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
	}

	if len(roleIDs) == 0 {
		return permissions, nil
	}

	// Get rules for these roles
	var rules []models.Rule
	if err := db.Table("rules").
		Joins("JOIN rule_roles ON rules.id = rule_roles.rule_id").
		Where("rule_roles.role_id IN ?", roleIDs).
		Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %v", err)
	}

	// Extract resource and action combinations
	for _, rule := range rules {
		permissions = append(permissions, fmt.Sprintf("%s %s", rule.Method, rule.Path))
	}

	return permissions, nil
}
