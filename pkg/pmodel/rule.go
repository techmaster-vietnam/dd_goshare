package pmodel

type Rule struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	Name       string `json:"name"`
	IsPrivate  bool   `json:"is_private"`
	AccessType int    `json:"access_type"` // 1: allow, 2: forbid, 3: allow_all, 4: forbid_all
	Roles      []int  `json:"roles"`       // Danh sách role IDs
}
