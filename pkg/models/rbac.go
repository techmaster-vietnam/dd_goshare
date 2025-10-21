package models

// Role model với các trường mở rộng cho RBAC hiện đại
// và quan hệ many2many với Rule
type Role struct {
	ID          int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"size:100;not null;unique;index" json:"name"`
	Description string `gorm:"size:255" json:"description,omitempty"`

	// Relationships
	Rules    []Rule `gorm:"many2many:rule_roles;" json:"rules,omitempty"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}

type Rule struct {
	ID         int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Path       string `gorm:"size:500;not null;uniqueIndex:idx_rule_unique" json:"path"`
	Method     string `gorm:"size:10;not null;uniqueIndex:idx_rule_unique" json:"method"`
	IsPrivate  bool   `gorm:"index" json:"is_private"`
	Service    string `gorm:"size:50;uniqueIndex:idx_rule_unique" json:"service"`
	// Relationships
	Roles []Role `gorm:"many2many:rule_roles;" json:"roles,omitempty"`
}

// TableName specifies the table name for Rule model
func (Rule) TableName() string {
	return "rules"
}

// UserRole liên kết user với nhiều role
type UserRole struct {
	UserID string `gorm:"primaryKey;size:50;index" json:"user_id"`
	RoleID int    `gorm:"primaryKey;index" json:"role_id"`

	// Relationships
	Role Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
}

// TableName specifies the table name for UserRole model
func (UserRole) TableName() string {
	return "user_roles"
}

// RuleRole liên kết rule với nhiều role, cho phép access_type riêng cho từng role trên từng rule
// Nếu access_type là NULL thì mặc định lấy theo rule
type RuleRole struct {
	RuleID     int     `gorm:"primaryKey;index" json:"rule_id"`
	RoleID     int     `gorm:"primaryKey;index" json:"role_id"`

	// Relationships
	Rule Rule `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE" json:"rule,omitempty"`
	Role Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
}

// TableName specifies the table name for RuleRole model
func (RuleRole) TableName() string {
	return "rule_roles"
}

// Helper methods for Role
func (r *Role) IsAdmin() bool {
	return r.Name == "admin"
}

// Helper methods for Rule
func (r *Rule) IsPublic() bool {
	return !r.IsPrivate
}

func (r *Rule) GetRouteKey() string {
	return r.Method + " " + r.Path
}


// RoleRequest dùng để validate và nhận dữ liệu tạo/cập nhật role từ API
type RoleRequest struct {
	Name        string `json:"name" valid:"required~Tên không được để trống, runelength(4|100)~Tên không hợp lệ (4-100 ký tự)"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"`
}

// SetRole chuyển RoleRequest thành Role
func (r *RoleRequest) SetRole() Role {
	return Role{
		Name:        r.Name,
		Description: r.Description,
	}
}
