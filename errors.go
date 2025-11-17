package client

import "errors"

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden: insufficient permissions to perform action")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidLogin         = errors.New("invalid login format")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInvalidChallenge     = errors.New("invalid challenge")
	ErrDiaryNotFound        = errors.New("diary not found")
	ErrEntryNotFound        = errors.New("entry not found")
	ErrTopicNotFound        = errors.New("topic not found")
	ErrTemplateNotFound     = errors.New("template not found")
	ErrDiaryLimitExceeded   = errors.New("diary limit exceeded")
)
