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

	t.Run("TestAllow", func(t *testing.T) {
		roleExp := Allow(1, 2) // admin, moderator
		roles, action := roleExp()

		if action != models.Allow {
			t.Errorf("Expected action %d, got %d", models.Allow, action)
		}

		if !roles[1].(bool) {
			t.Error("Admin role should be allowed")
		}

		if !roles[2].(bool) {
			t.Error("Moderator role should be allowed")
		}

		if _, exists := roles[3]; exists {
			t.Error("User role should not be in allow list")
		}
	})

	t.Run("TestAllowAll", func(t *testing.T) {
		roleExp := AllowAll()
		roles, action := roleExp()

		if action != models.AllowAll {
			t.Errorf("Expected action %d, got %d", models.AllowAll, action)
		}

		expectedRoles := []int{1, 2, 3, 4}
		for _, roleID := range expectedRoles {
			if !roles[roleID].(bool) {
				t.Errorf("Role %d should be allowed", roleID)
			}
		}
	})

	t.Run("TestForbid", func(t *testing.T) {
		roleExp := Forbid(4) // guest
		roles, action := roleExp()

		if action != models.Forbid {
			t.Errorf("Expected action %d, got %d", models.Forbid, action)
		}

		if roles[4].(bool) {
			t.Error("Guest role should be forbidden (false)")
		}

		if _, exists := roles[1]; exists {
			t.Error("Admin role should not be in forbid list")
		}
	})

	t.Run("TestForbidAll", func(t *testing.T) {
		roleExp := ForbidAll()
		roles, action := roleExp()

		if action != models.ForbidAll {
			t.Errorf("Expected action %d, got %d", models.ForbidAll, action)
		}

		expectedRoles := []int{1, 2, 3, 4}
		for _, roleID := range expectedRoles {
			if roles[roleID].(bool) {
				t.Errorf("Role %d should be forbidden (false)", roleID)
			}
		}
	})
}

func TestRoleExpCombination(t *testing.T) {
	// Setup test roles
	Roles = map[string]int{
		"admin":     1,
		"moderator": 2,
		"user":      3,
		"guest":     4,
	}

	t.Run("TestComplexAllow", func(t *testing.T) {
		// Chỉ cho phép admin và user
		roleExp := Allow(1, 3)
		roles, action := roleExp()

		if action != models.Allow {
			t.Errorf("Expected action %d, got %d", models.Allow, action)
		}

		// Check allowed roles
		allowedRoles := []int{1, 3}
		for _, roleID := range allowedRoles {
			if !roles[roleID].(bool) {
				t.Errorf("Role %d should be allowed", roleID)
			}
		}

		// Check not allowed roles
		notAllowedRoles := []int{2, 4}
		for _, roleID := range notAllowedRoles {
			if _, exists := roles[roleID]; exists {
				t.Errorf("Role %d should not be in allow map", roleID)
			}
		}
	})

	t.Run("TestComplexForbid", func(t *testing.T) {
		// Cấm guest và user
		roleExp := Forbid(3, 4)
		roles, action := roleExp()

		if action != models.Forbid {
			t.Errorf("Expected action %d, got %d", models.Forbid, action)
		}

		// Check forbidden roles
		forbiddenRoles := []int{3, 4}
		for _, roleID := range forbiddenRoles {
			if roles[roleID].(bool) {
				t.Errorf("Role %d should be forbidden (false)", roleID)
			}
		}

		// Check roles not in forbid list
		notForbiddenRoles := []int{1, 2}
		for _, roleID := range notForbiddenRoles {
			if _, exists := roles[roleID]; exists {
				t.Errorf("Role %d should not be in forbid map", roleID)
			}
		}
	})
}
