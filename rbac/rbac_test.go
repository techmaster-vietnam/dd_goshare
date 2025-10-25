package rbac

import (
	"testing"

	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
)

func TestRBACFunctions(t *testing.T) {
	// Setup test roles
	Roles = map[string]int{
		"admin":     1,
		"moderator": 2,
		"user":      3,
		"guest":     4,
	}

	t.Run("TestAllowProtected", func(t *testing.T) {
		roleExp := AllowProtected(1, 2) // admin, moderator
		roles, action := roleExp()
		if action != models.Protected {
			t.Errorf("Expected action %d, got %d", models.Protected, action)
		}
		if b, ok := roles[1].(*bool); !ok || b == nil || !*b {
			t.Error("Admin role should be allowed (true)")
		}
		if b, ok := roles[2].(*bool); !ok || b == nil || !*b {
			t.Error("Moderator role should be allowed (true)")
		}
		if b, exists := roles[3]; exists && b != nil {
			t.Error("User role should not be in allow list")
		}
	})

	t.Run("TestAllowAllRoute", func(t *testing.T) {
		route := models.Rule{AccessType: models.AllowAll, IsPrivate: true}
		if !route.IsPrivate {
			t.Error("AllowAll route should be private (require login)")
		}
		if route.AccessType != models.AllowAll {
			t.Errorf("Expected access_type AllowAll, got %d", route.AccessType)
		}
	})

	t.Run("TestForbidAllRoute", func(t *testing.T) {
		route := models.Rule{AccessType: models.ForbidAll, IsPrivate: true}
		if !route.IsPrivate {
			t.Error("ForbidAll route should be private (require login)")
		}
		if route.AccessType != models.ForbidAll {
			t.Errorf("Expected access_type ForbidAll, got %d", route.AccessType)
		}
	})

	t.Run("TestPublicRoute", func(t *testing.T) {
		route := models.Rule{AccessType: models.AllowAll, IsPrivate: false}
		if route.IsPrivate {
			t.Error("Public route should not require login (is_private=false)")
		}
		if route.AccessType != models.AllowAll {
			t.Errorf("Expected access_type AllowAll, got %d", route.AccessType)
		}
	})
}

func TestRoleExpCombination(t *testing.T) {
	t.Run("TestForbidAllAccess", func(t *testing.T) {
		route := models.Rule{AccessType: models.ForbidAll, IsPrivate: true}
		// Giả lập middleware: đã đăng nhập nhưng bị cấm
		if route.AccessType == models.ForbidAll && route.IsPrivate {
			// Không cho phép truy cập
		} else {
			t.Error("ForbidAll route logic failed")
		}
	})

	t.Run("TestImplicitDeny", func(t *testing.T) {
		roleExp := AllowProtected(1)
		roles, _ := roleExp()
		if _, exists := roles[3]; exists {
			t.Error("Role 3 should not be in allow map (implicit deny)")
		}
	})

	t.Run("TestComplexAllowProtected", func(t *testing.T) {
		roleExp := AllowProtected(1, 3)
		roles, action := roleExp()

		if action != models.Protected {
			t.Errorf("Expected action %d, got %d", models.Protected, action)
		}

		allowedRoles := []int{1, 3}
		for _, roleID := range allowedRoles {
			if b, ok := roles[roleID].(*bool); !ok || b == nil || !*b {
				t.Errorf("Role %d should be allowed (true)", roleID)
			}
		}

		notAllowedRoles := []int{2, 4}
		for _, roleID := range notAllowedRoles {
			if b, exists := roles[roleID]; exists && b != nil {
				t.Errorf("Role %d should not be in allow map", roleID)
			}
		}
	})

}
