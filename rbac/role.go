package rbac

import "github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"

const (
	ALLOW      = "allow"
	ALLOW_ALL  = "allow_all"
	FORBID     = "forbid"
	FORBID_ALL = "forbid_all"
)

// Cho phép 1 hoặc nhiều role truy cập
func Allow(roles ...int) RoleExp {
	return func() (pmodel.Roles, string) {
		mapRoles := make(pmodel.Roles)
		for _, role := range roles {
			mapRoles[role] = true
		}
		return mapRoles, ALLOW
	}
}

// Cho phép tất cả các role
func AllowAll() RoleExp {
	return func() (pmodel.Roles, string) {
		mapRoles := make(pmodel.Roles)
		for _, roleID := range Roles {
			mapRoles[roleID] = true
		}
		return mapRoles, ALLOW_ALL
	}
}

// Cấm 1 hoặc nhiều role truy cập
func Forbid(roles ...int) RoleExp {
	return func() (pmodel.Roles, string) {
		mapRoles := make(pmodel.Roles)
		for _, role := range roles {
			mapRoles[role] = false
		}
		return mapRoles, FORBID
	}
}

// Cấm tất cả các role (trừ root nếu muốn tuỳ chỉnh)
func ForbidAll() RoleExp {
	return func() (pmodel.Roles, string) {
		mapRoles := make(pmodel.Roles)
		for _, roleID := range Roles {
			mapRoles[roleID] = false
		}
		return mapRoles, FORBID_ALL
	}
}
