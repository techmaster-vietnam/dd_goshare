package pmodel

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Roles là map[int]interface{} để lưu role ID và trạng thái allow/forbid
type Roles map[int]interface{}
