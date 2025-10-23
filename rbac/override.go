package rbac

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

// GetRuleRoleAccessType returns the access_type override for a given rule_id and role_id, or nil if not set
func GetRuleRoleAccessType(db *gorm.DB, ruleID int, roleID int) *int {
	var rr models.RuleRole
	err := db.Table("rule_roles").Where("rule_id = ? AND role_id = ?", ruleID, roleID).First(&rr).Error
	if err == nil && rr.AccessType != nil {
		return rr.AccessType
	}
	return nil
}
