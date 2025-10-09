package rbac

import (
	"fmt"
	"strings"
)

// DebugRouteRole in ra thông tin Private Route - Role theo Core pattern
func DebugRouteRole() {
	fmt.Println("*** Private Routes ***")
	fmt.Printf("Total routes: %d\n", len(routesRoles))

	for routeKey, route := range routesRoles {
		if route.IsPrivate {
			fmt.Printf("- %s (Private)\n", routeKey)
			for roleID, allow := range route.Roles {
				roleName := getRoleName(roleID)
				if allow.(bool) {
					fmt.Printf("      ✅ %s (allow)\n", roleName)
				} else {
					fmt.Printf("      ❌ %s (forbid)\n", roleName)
				}
			}
		}
	}
}

// DebugPublicRoutes in ra danh sách những đường dẫn public không kiểm tra quyền
func DebugPublicRoutes() {
	fmt.Println("*** Public Routes ***")
	fmt.Printf("Total: %d\n", len(publicRoutes))

	for route := range publicRoutes {
		fmt.Printf("- %s\n", route)
	}
}

// DebugPathRole in ra thông tin debug Route - Role thành 2 phần
func DebugPathRole() {
	fmt.Println("*** Routes by Path ***")

	for path, route := range pathsRoles {
		fmt.Printf("- %s (%s)\n", path, route.Method)
		for roleID := range route.Roles {
			roleName := getRoleName(roleID)
			fmt.Printf("     %s\n", roleName)
		}
	}
}

// DebugUserPermissions debug specific user permissions
func DebugUserPermissions(userID string) {
	fmt.Printf("*** User Permissions: %s ***\n", userID)

	// Get user roles from database
	db := GetDB()
	if db == nil {
		fmt.Println("Database not initialized")
		return
	}

	type UserRoleInfo struct {
		RoleID   int    `json:"role_id"`
		RoleName string `json:"role_name"`
	}

	var userRoles []UserRoleInfo
	err := db.Table("user_roles ur").
		Select("ur.role_id, r.name as role_name").
		Joins("JOIN roles r ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Find(&userRoles).Error

	if err != nil {
		fmt.Printf("Error loading user roles: %v\n", err)
		return
	}

	if len(userRoles) == 0 {
		fmt.Println("User has no roles assigned")
		return
	}

	fmt.Printf("User roles (%d):\n", len(userRoles))
	for _, role := range userRoles {
		fmt.Printf("  - %s (ID: %d)\n", role.RoleName, role.RoleID)
	}

	// Check permissions for each route
	fmt.Println("\nRoute permissions:")
	userRoleMap := make(map[int]bool)
	for _, role := range userRoles {
		userRoleMap[role.RoleID] = true
	}

	for routeKey, route := range routesRoles {
		if route.IsPrivate {
			hasPermission := checkUserRouteRoleIntersect(userRoleMap, route.Roles)
			status := "❌ DENIED"
			if hasPermission {
				status = "✅ ALLOWED"
			}
			fmt.Printf("  %s %s\n", status, routeKey)
		}
	}
}

// DebugRoleHierarchy in ra hierarchy của roles nếu có parent-child relationship
func DebugRoleHierarchy() {
	fmt.Println("*** Role Hierarchy ***")

	db := GetDB()
	if db == nil {
		fmt.Println("Database not initialized")
		return
	}

	type RoleHierarchy struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		ParentID   *int   `json:"parent_id"`
		ParentName string `json:"parent_name"`
	}

	var roles []RoleHierarchy
	err := db.Table("roles r").
		Select("r.id, r.name, r.parent_id, p.name as parent_name").
		Joins("LEFT JOIN roles p ON r.parent_id = p.id").
		Order("r.parent_id, r.name").
		Find(&roles).Error

	if err != nil {
		fmt.Printf("Error loading role hierarchy: %v\n", err)
		return
	}

	// Group by parent
	parentMap := make(map[string][]RoleHierarchy)
	for _, role := range roles {
		parentKey := "ROOT"
		if role.ParentID != nil {
			parentKey = role.ParentName
		}
		parentMap[parentKey] = append(parentMap[parentKey], role)
	}

	// Print hierarchy
	for parent, children := range parentMap {
		if parent == "ROOT" {
			fmt.Println("Root roles:")
		} else {
			fmt.Printf("Under %s:\n", parent)
		}

		for _, child := range children {
			indent := "  "
			if parent != "ROOT" {
				indent = "    "
			}
			fmt.Printf("%s- %s (ID: %d)\n", indent, child.Name, child.ID)
		}
		fmt.Println()
	}
}

// DebugSystemInfo in ra thông tin tổng quan về hệ thống RBAC
func DebugSystemInfo() {
	fmt.Println("*** RBAC System Information ***")
	fmt.Printf("Service: %s\n", config.Service)
	fmt.Printf("Highest Role: %s\n", config.HighestRole)
	fmt.Printf("Make Unassigned Route Public: %t\n", config.MakeUnassignedRoutePublic)
	fmt.Printf("Total Roles in Memory: %d\n", len(Roles))
	fmt.Printf("Total Routes: %d\n", len(routesRoles))
	fmt.Printf("Total Public Routes: %d\n", len(publicRoutes))
	fmt.Printf("Total Paths: %d\n", len(pathsRoles))

	fmt.Println("\nRoles in memory:")
	for roleName, roleID := range Roles {
		fmt.Printf("  - %s (ID: %d)\n", roleName, roleID)
	}
}

// getRoleName helper function to get role name by ID
func getRoleName(roleID int) string {
	if name, exists := roleName[roleID]; exists {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", roleID)
}

// correctRoute inserts space between HTTP Verb and Path (from Core pattern)
func correctRoute(route string) string {
	posFirstSlash := strings.Index(route, "/")
	if posFirstSlash == -1 {
		return route
	}
	return route[0:posFirstSlash] + " " + route[posFirstSlash:]
}

// RoleName chuyển role từ int thành string (from Core pattern)
func RoleName(roleID int) string {
	return getRoleName(roleID)
}

// RoleNames chuyển roles kiểu map[int]bool thành mảng string mô tả các role
func RoleNames(roles map[int]bool) []string {
	var roleNames []string
	for roleID := range roles {
		roleNames = append(roleNames, getRoleName(roleID))
	}
	return roleNames
}
