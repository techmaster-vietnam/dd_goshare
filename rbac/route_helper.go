package rbac

import (
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
)

// RoleExp là biểu thức phân quyền động theo Core pattern
type RoleExp func() (pmodel.Roles, int)

// assignRoles gán role vào route và path theo Core pattern
func assignRoles(route Route) {
	re, _ := regexp.Compile("/+")
	route.Path = re.ReplaceAllLiteralString(route.Path, "/")

	routeKey := route.Method + " " + route.Path
	routesRoles[routeKey] = route

	// ✅ IMPORTANT: Add to freshRoutes so RegisterRulesToDB can sync to DB
	freshRoutes[routeKey] = route

	if _, ok := pathsRoles[route.Path]; !ok {
		pathsRoles[route.Path] = route
	}

	if !route.IsPrivate {
		publicRoutes[routeKey] = true
	}
}

func getFullPath(_ fiber.Router, path string) string {
	if !strings.HasPrefix(path, "/api") && strings.HasPrefix(path, "/") {
		return "/api" + path
	}
	return path
}

// Get đăng ký GET route với RBAC v2.0, isPrivate và accessType hoàn toàn độc lập
func Get(group fiber.Router, path string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "GET",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
	}
	if isPrivate {
		log.Printf("DEBUG ROUTE: Registering private GET route %s with RBAC middleware", path)
		group.Get(path, CheckPermissionMiddleware(), handler)
	} else {
		log.Printf("DEBUG ROUTE: Registering public/protected GET route %s", path)
		group.Get(path, handler)
	}
	assignRoles(route)
}

func Post(group fiber.Router, path string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "POST",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
	}
	if isPrivate {
		group.Post(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Post(path, handler)
	}
	assignRoles(route)
}

func Put(group fiber.Router, path string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "PUT",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
	}
	if isPrivate {
		group.Put(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Put(path, handler)
	}
	assignRoles(route)
}

func Delete(group fiber.Router, path string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "DELETE",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
	}
	if isPrivate {
		group.Delete(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Delete(path, handler)
	}
	assignRoles(route)
}

func Patch(group fiber.Router, path string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	route := Route{
		Path:       getFullPath(group, path),
		Method:     "PATCH",
		IsPrivate:  isPrivate,
		Roles:      roles,
		AccessType: accessType,
	}
	if isPrivate {
		group.Patch(path, CheckPermissionMiddleware(), handler)
	} else {
		group.Patch(path, handler)
	}
	assignRoles(route)
}

func Any(group fiber.Router, path string, businessName string, isPrivate bool, roleExp RoleExp, handler fiber.Handler) {
	roles, accessType := roleExp()
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		route := Route{
			Path:       getFullPath(group, path),
			Method:     method,
			IsPrivate:  isPrivate,
			Roles:      roles,
			AccessType: accessType,
		}
		assignRoles(route)
	}
	group.All(path, handler)
}
