package pmodel

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Roles []Role `json:"roles"`
}
