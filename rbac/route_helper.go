package rbac

import (
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
)

// RoleExp là biểu thức phân quyền động theo Core pattern
type RoleExp func() (pmodel.Roles, string)

// assignRoles gán role vào route và path theo Core pattern
func assignRoles(route Route) {
	// Sử dụng regex để thay thế double slashes bằng single slash
	re, _ := regexp.Compile("/+")
	route.Path = re.ReplaceAllLiteralString(route.Path, "/")

	routeKey := route.Method + " " + route.Path
	routesRoles[routeKey] = route

	// Store by path for debugging
	if _, ok := pathsRoles[route.Path]; !ok {
		pathsRoles[route.Path] = route
	}

	// Track public routes
	if !route.IsPrivate {
		publicRoutes[routeKey] = true
	}
}

// getFullPath constructs full path from fiber router group
func getFullPath(_ fiber.Router, path string) string {
	// For API routes, add the /api prefix if not already present
	if !strings.HasPrefix(path, "/api") && strings.HasPrefix(path, "/") {
		return "/api" + path
	}
	return path
}

// Get đăng ký GET route với RBAC theo Core pattern
func Get(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "GET",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
		Name:       businessName,
	}

	if isPrivate {
		log.Printf("DEBUG ROUTE: Registering private GET route %s with RBAC middleware", path)
		group.Get(path, CheckPermissionMiddleware(), handler)
	} else {
		log.Printf("DEBUG ROUTE: Registering public GET route %s", path)
		group.Get(path, handler)
	}
	assignRoles(route)
}

// Post đăng ký POST route với RBAC theo Core pattern
func Post(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "POST",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
		Name:       businessName,
	}

	if isPrivate {
		group.Post(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Post(path, handler)
	}
	assignRoles(route)
}

// Put đăng ký PUT route với RBAC theo Core pattern
func Put(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "PUT",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
		Name:       businessName,
	}

	if isPrivate {
		group.Put(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Put(path, handler)
	}
	assignRoles(route)
}

// Delete đăng ký DELETE route với RBAC theo Core pattern
func Delete(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "DELETE",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
		Name:       businessName,
	}

	if isPrivate {
		group.Delete(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Delete(path, handler)
	}
	assignRoles(route)
}

// Patch đăng ký PATCH route với RBAC theo Core pattern
func Patch(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "PATCH",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
		Name:       businessName,
	}

	group.Patch(path, handler)
	assignRoles(route)
}

// Any đăng ký tất cả HTTP methods với RBAC theo Core pattern
func Any(group fiber.Router, path string, businessName string, roleExp RoleExp, isPrivate bool, handler fiber.Handler) {
	roles, accessType := roleExp()

	// List of HTTP methods to register
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		route := Route{
			Path:       getFullPath(group, path),
			Method:     method,
			IsPrivate:  isPrivate,
			Roles:      roles,
			AccessType: accessType,
			Name:       businessName,
		}
		assignRoles(route)
	}
	// getBusinessFunctionName extracts the business function name from roles map if present
	// The function getBusinessFunctionName is no longer needed and has been removed.
	// func getBusinessFunctionName(roles pmodel.Roles) string {
	//       return ""
	// }

	group.All(path, handler)
}
