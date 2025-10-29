package models

type Comment struct {
	ID              string  `gorm:"primaryKey;size:12" json:"id"`
	DialogID        string  `gorm:"size:12;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"dialog_id"`
	UserID          string  `gorm:"size:50;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user_id"`
	Content         string  `gorm:"type:text" json:"content"`
	Likes           int     `gorm:"default:0" json:"likes"`
	Rating          *int    `json:"rating"`
	ParentCommentID *string `gorm:"size:12;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"parent_comment_id"`

	// Quan hệ với các bảng khác
	Dialog  Dialog    `gorm:"foreignKey:DialogID;references:ID" json:"-"`
	User    Customer  `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Parent  *Comment  `gorm:"foreignKey:ParentCommentID;references:ID" json:"-"`
	Replies []Comment `gorm:"foreignKey:ParentCommentID;references:ID" json:"-"`
}

// TableName chỉ định tên bảng cho Comment
func (Comment) TableName() string {
	return "comments"
}

type CreateCommentRequest struct {
	DialogID        string `json:"dialog_id"`
	ParentCommentID string `json:"parent_comment_id"`
	Content         string `json:"content"`
	Rating          int    `json:"rating"`
}

type UpdateCommentRequest struct {
	CommentID string `json:"comment_id"`
	Content   string `json:"content"`
	Rating    int    `json:"rating"`
	Likes     int    `json:"likes"`
}
type CreateCommentResponse struct {
	StatusCode int     `json:"status_code"`
	Message    string  `json:"message"`
	Comment    Comment `json:"comment"`
}
type CommentResponse struct {
	UserName   string  `json:"user_name"`
	UserAvatar string  `json:"user_avatar"`
	Comment    Comment `json:"comment"`
}
type GetCommentsByDialogIDResponse struct {
	StatusCode int       `json:"status_code"`
	Message    string    `json:"message"`
	Comments   []Comment `json:"comments"`
}
