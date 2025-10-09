package services

import "errors"

// Service-level errors
var (
	ErrTopicNameRequired = errors.New("topic name is required")
	ErrTopicIDRequired   = errors.New("topic ID is required")
	ErrTopicNotFound     = errors.New("topic not found")
	ErrTopicExists       = errors.New("topic already exists")

	// Dialog errors
	ErrDialogRawTextRequired = errors.New("dialog raw text is required")
	ErrDialogScriptRequired  = errors.New("dialog script is required")
	ErrDialogTitleRequired   = errors.New("dialog title is required")
	ErrDialogLevelIDRequired = errors.New("dialog level ID is required")
	ErrDialogIDRequired      = errors.New("dialog ID is required")
	ErrDialogNotFound        = errors.New("dialog not found")

	// Audio errors
	ErrAudioPathRequired = errors.New("audio path is required")
	ErrAudioTypeRequired = errors.New("audio type is required")
	ErrAudioIDRequired   = errors.New("audio ID is required")
	ErrAudioNotFound     = errors.New("audio not found")

	// Level errors
	ErrLevelNameRequired = errors.New("level name is required")
	ErrLevelIDRequired   = errors.New("level ID is required")
	ErrLevelNotFound     = errors.New("level not found")

	// User errors
	ErrEmailRequired     = errors.New("email is required")
	ErrPasswordRequired  = errors.New("password is required")
	ErrUsernameRequired  = errors.New("username is required")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserData   = errors.New("invalid user data")
	ErrInvalidPassword   = errors.New("invalid password")

	// Comment errors
	ErrParentCommentIDRequired       = errors.New("parent comment ID is required")
	ErrParentCommentDialogIDNotMatch = errors.New("parent comment dialog ID does not match")

	// Tag errors
	ErrTagNameRequired  = errors.New("tag name is required")
	ErrTagNotFound      = errors.New("tag not found")
	ErrTagAlreadyExists = errors.New("tag already exists")
)
