package rbac

import (
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
)

func AllowProtected(roles ...int) RoleExp {
	return func() (pmodel.Roles, int) {
		mapRoles := make(pmodel.Roles)
		for _, role := range roles {
			b := true
			mapRoles[role] = &b
		}
		return mapRoles, models.Protected
	}
}

// Public route (ai cũng truy cập được)
func PublicRoute() RoleExp {
	return func() (pmodel.Roles, int) {
		return make(pmodel.Roles), models.AllowAll
	}
}

// Private route (chỉ admin)
func PrivateRoute() RoleExp {
	return func() (pmodel.Roles, int) {
		return make(pmodel.Roles), models.ForbidAll
	}
}
