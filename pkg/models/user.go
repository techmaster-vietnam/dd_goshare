package models

// User dùng chung cho customer/employee
type User struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	Password    string     `json:"password"`
	AvatarURL   string     `json:"avatar_url"`
	Streak      int        `json:"streak"`
	Score       int        `json:"score"`
}