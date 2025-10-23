package rbac

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
)

// Cho phép 1 hoặc nhiều role truy cập
func Allow(roles ...int) RoleExp {
	return func() (pmodel.Roles, int) {
		mapRoles := make(pmodel.Roles)
		for _, role := range roles {
			mapRoles[role] = true
		}
		return mapRoles, models.Allow
	}
}

// Cho phép tất cả các role
func AllowAll() RoleExp {
	return func() (pmodel.Roles, int) {
		mapRoles := make(pmodel.Roles)
		for _, roleID := range Roles {
			mapRoles[roleID] = true
		}
		return mapRoles, models.AllowAll
	}
}

// Cấm 1 hoặc nhiều role truy cập
func Forbid(roles ...int) RoleExp {
	return func() (pmodel.Roles, int) {
		mapRoles := make(pmodel.Roles)
		for _, role := range roles {
			mapRoles[role] = false
		}
		return mapRoles, models.Forbid
	}
}

// Cấm tất cả các role (trừ root nếu muốn tuỳ chỉnh)
func ForbidAll() RoleExp {
	return func() (pmodel.Roles, int) {
		mapRoles := make(pmodel.Roles)
		for _, roleID := range Roles {
			mapRoles[roleID] = false
		}
		return mapRoles, models.ForbidAll
	}
}
