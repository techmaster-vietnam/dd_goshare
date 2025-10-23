package rbac

import (
	"fmt"
	"log"

	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/gorm"
)

// RefreshRoles reloads roles from the database (stub for now)
func RefreshRoles() error {
	// TODO: implement if needed
	return nil
}

// RefreshRules reloads rules from the database
func RefreshRules() error {
	return LoadRulesFromDB() // ✅ Gọi hàm đúng để update routesRoles
}

var db *gorm.DB

// SetDB sets the database instance for RBAC
func SetDB(database *gorm.DB) {
	log.Printf("SetDB called, db pointer: %v", database)
	db = database
}

// GetDB returns the current DB instance
func GetDB() *gorm.DB {
	log.Printf("GetDB called, db pointer: %v", db)
	return db
}

// LoadRules loads rules and their role mappings from database
func LoadRules() error {
	if db == nil {
		return fmt.Errorf("database not set")
	}

	// Query rules with their associated roles
	var results []struct {
		models.Rule
		RoleName string `gorm:"column:role_name"`
	}

	err := db.Table("rules").
		Select("rules.*, roles.name as role_name").
		Joins("LEFT JOIN rule_roles ON rules.id = rule_roles.rule_id").
		Joins("LEFT JOIN roles ON rule_roles.role_id = roles.id").
		Find(&results).Error

	if err != nil {
		return err
	}

	// Group by rule and build permission map
	ruleRoles := make(map[string]map[string][]string) // [path][method] = []roles

	for _, r := range results {
		path := r.Path
		method := r.Method

		if ruleRoles[path] == nil {
			ruleRoles[path] = make(map[string][]string)
		}
		if ruleRoles[path][method] == nil {
			ruleRoles[path][method] = []string{}
		}

		if r.RoleName != "" {
			ruleRoles[path][method] = append(ruleRoles[path][method], r.RoleName)
		}
	}

	// Rules are now loaded directly in rbac.go through LoadRulesFromDB()
	// This function is kept for compatibility but does nothing

	return nil
}
