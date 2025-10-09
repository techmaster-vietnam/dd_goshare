package pmodel

type Rule struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	Name       string `json:"name"`
	IsPrivate  bool   `json:"is_private"`
	AccessType string `json:"access_type"` // allow, forbid, allow_all, forbid_all
	Roles      []int  `json:"roles"`       // Danh sách role IDs
}
