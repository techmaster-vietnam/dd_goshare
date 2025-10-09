package models

type Employee struct {
	ID       string `gorm:"primaryKey;size:12" json:"id"`
	Name     string `gorm:"size:50" json:"name"`
	Email    string `gorm:"size:100;uniqueIndex" json:"email"`
	Password string `gorm:"size:255" json:"-"`
	// Không lưu trực tiếp role ở đây, dùng bảng UserRole để phân quyền
}

func (Employee) TableName() string {
	return "employees"
}
