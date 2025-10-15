package rbac

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
)

// Business function mappings - đây là các chức năng thực tế của hệ thống
const (
	// Auth functions
	LOGIN    = "auth.login"    // POST /auth/login
	REGISTER = "auth.register" // POST /auth/register

	// Role management functions
	VIEW_ROLES_RULE   = "roles.rule.view"   // GET /roles, GET /users/:userId/roles, GET /roles/:roleId/rules, GET /rules/:ruleId/roles
	VIEW_RULES_ROLE   = "rule.roles.view"   // GET /rules, GET /users/:userId/roles, GET /roles/:roleId/rules, GET /rules/:ruleId/roles
	VIEW_ROLES        = "role.view"         // GET /roles
	CREATE_ROLES = "role.create" // POST /roles
	UPDATE_ROLES = "role.update" // PUT /roles/:id
	DELETE_ROLES = "role.delete" // DELETE /roles/:id
	ASSIGN_ROLES = "role.assign" // POST/DELETE/PUT /roles/:roleId/users/:userId, /users/:userId/roles, /roles/:roleId/rules, /rules/:ruleId/roles

	// Rule management functions
	VIEW_RULES   = "rule.view"   // GET /rules
	CREATE_RULES = "rule.create" // POST /rules

	// Dialog management functions
	VIEW_DIALOG_RESULT      = "dialog.view.result"      // GET /dialog/:dialogID
	UPDATE_DIALOG_RESULT    = "dialog.update.result"    // PUT /dialog/:dialogID
	UPDATE_DIALOG_IMAGES    = "dialog.update.images"    // POST /dialog/dialogs/:dialogID/images
	SERVE_DIALOG_AUDIO_FILE = "dialog.serve.audio"      // GET /dialog/audio/:filename
	VIEW_ALL_DIALOG_RESULTS = "dialog.view.all.results" // GET /results
	VIEW_ALL_DIALOGS        = "dialog.view.all"         // GET /dialogs

	// Word management functions
	CREATE_FILLWORDS = "fillword.create" // GET /generate/:dialogID

	// Image/Media management functions
	CREATE_IMAGE_FIGURE = "image.create.figure" // POST /figure
	VIEW_IMAGE_FIGURE    = "image.view.figure"   // GET /figure/:topicID
	UPLOAD_IMAGE        = "image.upload"        // POST /
	SERVE_IMAGE_FILE    = "image.serve.file"    // GET /:dialogID/:filename

	// Topic management functions
	CREATE_TOPICS = "topic.create" // POST /topic

	// Customer management functions
	VIEW_CUSTOMERS = "customer.view" // GET /customers

	// Subscription management functions
	CREATE_SUBSCRIPTION    = "subscription.create"      // POST /subscription
	UPDATE_SUBSCRIPTION    = "subscription.update"      // PUT /subscriptions/:id
	DELETE_SUBSCRIPTION    = "subscription.delete"      // DELETE /subscriptions/:id
	VIEW_ALL_SUBSCRIPTIONS = "subscription.view.all"    // GET /subscriptions
	VIEW_SUBSCRIPTION      = "subscription.view.single" // GET /subscriptions/:id

	// Employee management functions
	CREATE_EMPLOYEES = "employee.create" // POST /employees
	VIEW_EMPLOYEES   = "employee.view"   // GET /employees
	// Processing functions
	PROCESS_ALIGNMENT = "processing.align" // POST /align
)

// RequireFunction creates a RoleExp that checks which roles can access a business function
// This is dynamic - roles are loaded from database
func RequireFunction(functionName string) RoleExp {
	return func() (pmodel.Roles, string) {
		// Query database to find which roles have access to this function
		allowedRoles := getRolesForFunction(functionName)

		mapRoles := make(pmodel.Roles)
		for roleID := range allowedRoles {
			mapRoles[roleID] = true
		}

		return mapRoles, functionName
	}
}

// getRolesForFunction queries database to find roles that have access to a function
func getRolesForFunction(functionName string) map[int]bool {
	database := GetDB()
	if database == nil {
		return make(map[int]bool)
	}

	log.Printf("DEBUG getRolesForFunction: functionName = '%s'", functionName)

	var roleIDs []int
	err := database.Raw(`
		SELECT DISTINCT r.id 
		FROM roles r
		JOIN rule_roles rr ON r.id = rr.role_id
		JOIN rules ru ON rr.rule_id = ru.id
		WHERE ru.access_type = ?
	`, functionName).Scan(&roleIDs).Error

	log.Printf("DEBUG getRolesForFunction: query error = %v, roleIDs = %v", err, roleIDs)

	if err != nil {
		return make(map[int]bool)
	}

	result := make(map[int]bool)
	for _, roleID := range roleIDs {
		result[roleID] = true
	}

	log.Printf("DEBUG getRolesForFunction: result map = %v", result)
	return result
}

// RegisterBusinessRoute đăng ký route với kiểm tra quyền từ database
func RegisterBusinessRoute(group fiber.Router, method, path, businessFunction string, isPrivate bool, handler fiber.Handler) {
	log.Printf("🔍 RegisterBusinessRoute called for %s %s (function: %s, private: %v)", method, path, businessFunction, isPrivate)
	log.Printf("🔍 Current Roles map at route registration: %v", Roles)

	exp := RequireFunction(businessFunction)

	// ✅ THÊM: Track fresh routes từ code
	serviceValue := "dd_backend" // Default fallback
	if config.Service != "" {
		serviceValue = config.Service
	}
	fullPath := getFullPath(group, path)
	key := method + "|" + fullPath + "|" + serviceValue
	freshRoutes[key] = Route{
		Method:     method,
		Path:       fullPath,
		AccessType: businessFunction,
		IsPrivate:  isPrivate,
	}

	// Vẫn thêm vào routesRoles để logic cũ hoạt động
	routeKey := method + " " + path
	routesRoles[routeKey] = Route{
		Method:     method,
		Path:       fullPath,
		AccessType: businessFunction,
		IsPrivate:  isPrivate,
		Roles:      make(pmodel.Roles), // Will be populated by RequireFunction
	}

	log.Printf("🔄 Registered fresh route: %s", key)

	switch strings.ToUpper(method) {
	case "GET":
		Get(group, path, businessFunction, exp, isPrivate, handler)
	case "POST":
		Post(group, path, businessFunction, exp, isPrivate, handler)
	case "PUT":
		Put(group, path, businessFunction, exp, isPrivate, handler)
	case "DELETE":
		Delete(group, path, businessFunction, exp, isPrivate, handler)
	default:
		panic(fmt.Sprintf("Unsupported HTTP method: %s", method))
	}
}
