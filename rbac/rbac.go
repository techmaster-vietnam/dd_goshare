package rbac

import (
	"fmt"
	"strings"

	"github.com/techmaster-vietnam/dd_goshare/pkg/pmodel"
	"gorm.io/gorm"
)

const (
	DEFAULT_HIGHEST_ROLE = "admin"
)

// Enhanced Config theo Core RBAC pattern
type Config struct {
	MakeUnassignedRoutePublic bool     // true: routes without rules are public
	Service                   string   // service name for multi-service rules
	HighestRole               string   // highest privilege role (default: admin)
	DefaultRoles              []string // roles to create if missing
	DatabaseAutoMigrate       bool     // auto migrate database
}

// NewConfig creates default RBAC configuration
func NewConfig() Config {
	return Config{
		MakeUnassignedRoutePublic: false,
		Service:                   "dd_backend",
		HighestRole:               DEFAULT_HIGHEST_ROLE,
		DatabaseAutoMigrate:       true,
	}
}

// Validate checks if config is valid
func (c *Config) Validate() error {
	if c.Service == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if c.HighestRole == "" {
		c.HighestRole = DEFAULT_HIGHEST_ROLE
	}
	return nil
}

// Global variables theo Core pattern
var (
	roleName                    = map[int]string{}
	Roles        map[string]int = map[string]int{}
	routesRoles                 = make(map[string]Route)
	freshRoutes                 = make(map[string]Route)
	pathsRoles                  = make(map[string]Route)
	publicRoutes                = make(map[string]bool)
	config       Config
)

// Cấu trúc dùng để lưu thông tin của một route
type Route struct {
	Path       string
	Method     string
	IsPrivate  bool
	Roles      pmodel.Roles
	AccessType string
	Name       string // Business function name (e.g., "dialog.create")
}

// Cấu trúc dùng để lưu thông tin của một rule
type Rule struct {
	Path       string
	Method     string
	IsPrivate  bool
	Name       string
	AccessType string
	Service    string
}

// InitRBAC khởi tạo hệ thống RBAC với cấu hình
func InitRBAC(db *gorm.DB, configs ...Config) error {
	SetDB(db)

	if len(configs) == 0 {
		config = NewConfig()
	} else {
		config = configs[0]
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid RBAC config: %w", err)
	}

	if err := LoadRolesFromDB(); err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	if err := LoadRulesFromDB(); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	return nil
}

// LoadRolesFromDB loads roles from database into memory
func LoadRolesFromDB() error {
	database := GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	type Role struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	var roles []Role
	if err := database.Table("roles").Select("id, name").Find(&roles).Error; err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	Roles = make(map[string]int)
	roleName = make(map[int]string)

	for _, role := range roles {
		name := strings.ToLower(role.Name)
		Roles[name] = role.ID
		roleName[role.ID] = name
	}

	return nil
}

// LoadRulesFromDB loads rules from database và convert thành routes
func LoadRulesFromDB() error {
	database := GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	type RuleWithRoles struct {
		ID         int    `json:"id"`
		Path       string `json:"path"`
		Method     string `json:"method"`
		IsPrivate  bool   `json:"is_private"`
		Service    string `json:"service"`
		AccessType string `json:"access_type"`
		RoleIDs    string `json:"role_ids"`
	}

	var rules []RuleWithRoles
	query := `
	SELECT r.id, r.path, r.method, r.is_private, r.service, r.access_type,
       STRING_AGG(rr.role_id::text, ',') as role_ids
	FROM rules r
	LEFT JOIN rule_roles rr ON r.id = rr.rule_id
	WHERE r.service = 'dd_backend' OR r.service = ''
	GROUP BY r.id, r.path, r.method, r.access_type
	`

	if err := database.Raw(query).Scan(&rules).Error; err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	routesRoles = make(map[string]Route)
	pathsRoles = make(map[string]Route)
	publicRoutes = make(map[string]bool)

	for _, rule := range rules {
		route := Route{
			Path:       rule.Path,
			Method:     strings.ToUpper(rule.Method),
			IsPrivate:  rule.IsPrivate,
			AccessType: rule.AccessType,
			Roles:      make(pmodel.Roles),
		}

		// Parse role IDs và add vào Roles map
		if rule.RoleIDs != "" {
			roleIDStrs := strings.Split(rule.RoleIDs, ",")
			for _, roleIDStr := range roleIDStrs {
				roleID := parseInt(roleIDStr)
				if roleID > 0 {
					// Với access_type = "allow": role được phép (true)
					// Với access_type = "forbid": role bị cấm (false)
					switch rule.AccessType {
					case "forbid":
						route.Roles[roleID] = false
					default: // "allow", "allow_all", "forbid_all" hoặc empty
						route.Roles[roleID] = true
					}
				}
			}
		}

		routeKey := route.Method + " " + route.Path
		routesRoles[routeKey] = route
		pathsRoles[route.Path] = route

		if !route.IsPrivate {
			publicRoutes[routeKey] = true
		}
	}

	return nil
}

// ClearFreshRoutes - chỉ clear fresh routes, không động vào routesRoles
func ClearFreshRoutes() {
	freshRoutes = make(map[string]Route)
}

// Helper function to parse int from string
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			break
		}
	}
	return result
}
