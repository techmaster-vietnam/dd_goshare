package rbac

import (
	"fmt"
)

// AuthInfo represents authenticated user information
type AuthInfo struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
}

// GetUserRolesFromDB retrieves user roles from database
func GetUserRolesFromDB(userID string) ([]string, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var results []struct {
		Name string `gorm:"column:name"`
	}

	err := db.Table("roles").
		Select("roles.name").
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	roles := make([]string, len(results))
	for i, result := range results {
		roles[i] = result.Name
	}

	return roles, nil
}
