package client

import "errors"

var (
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden: insufficient permissions to perform action")
	ErrBadRequest           = errors.New("bad request")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrInvalidChallenge     = errors.New("invalid challenge")
	ErrDiaryNotFound        = errors.New("diary not found")
	ErrEntryNotFound        = errors.New("entry not found")
	ErrTopicNotFound        = errors.New("topic not found")
	ErrTemplateNotFound     = errors.New("template not found")
	ErrEntityRemoved        = errors.New("entity permanently removed")
	ErrEntityNotDeleted     = errors.New("entity is not deleted")
	ErrDiaryLimitExceeded   = errors.New("diary limit exceeded")
)
