package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/centrifugal/centrifuge-go"
	v1 "github.com/thingsdiary/client-go/proto"
)

const (
	realtimeCodeInvalidArgument    = 1000
	realtimeCodeUnauthenticated    = 1001
	realtimeCodePermissionDenied   = 1003
	realtimeCodeNotFound           = 1004
	realtimeCodeAlreadyExists      = 1005
	realtimeCodeConflict           = 1006
	realtimeCodeFailedPrecondition = 1007
	realtimeCodeResourceExhausted  = 1008
	realtimeCodeUnknownMethod      = 1010
	realtimeCodeInternal           = 1500
)

func mapRealtimeRPCError(method string, err error) error {
	var rpcErr *centrifuge.Error
	if !errors.As(err, &rpcErr) {
		return err
	}

	switch rpcErr.Code {
	case realtimeCodeUnauthenticated:
		return ErrUnauthorized
	case realtimeCodePermissionDenied:
		return ErrForbidden
	case realtimeCodeNotFound:
		switch {
		case strings.Contains(method, ".diary_key."):
			return ErrDiaryNotFound
		case strings.Contains(method, ".diary."):
			return ErrDiaryNotFound
		case strings.Contains(method, ".entry."):
			return ErrEntryNotFound
		case strings.Contains(method, ".topic."):
			return ErrTopicNotFound
		case strings.Contains(method, ".template."):
			return ErrTemplateNotFound
		}
	case realtimeCodeResourceExhausted:
		return ErrDiaryLimitExceeded
	case realtimeCodeFailedPrecondition:
		return ErrEntityRemoved
	case realtimeCodeInvalidArgument:
		return ErrBadRequest
	case realtimeCodeAlreadyExists, realtimeCodeConflict, realtimeCodeUnknownMethod, realtimeCodeInternal:
		// Keep the server message available to callers for errors without an
		// equivalent in the existing public client API.
	}

	return fmt.Errorf("realtime RPC failed (%d): %s", rpcErr.Code, rpcErr.Message)
}

func mapDiaryErrorCode(code v1.ErrorCode) error {
	switch code {
	case v1.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return nil
	case v1.ErrorCode_ERROR_CODE_BAD_DATA:
		return ErrBadRequest
	case v1.ErrorCode_ERROR_CODE_DIARIES_LIMIT_REACHED:
		return ErrDiaryLimitExceeded
	case v1.ErrorCode_ERROR_CODE_NOT_FOUND:
		return ErrDiaryNotFound
	default:
		return fmt.Errorf("diary operation failed: %s", code)
	}
}

func mapEntryErrorCode(code v1.ErrorCode) error {
	switch code {
	case v1.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return nil
	case v1.ErrorCode_ERROR_CODE_BAD_DATA:
		return ErrBadRequest
	default:
		return fmt.Errorf("entry operation failed: %s", code)
	}
}

func mapTopicErrorCode(code v1.ErrorCode) error {
	switch code {
	case v1.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return nil
	case v1.ErrorCode_ERROR_CODE_BAD_DATA:
		return ErrBadRequest
	case v1.ErrorCode_ERROR_CODE_FAILED_PRECONDITION:
		return ErrTopicEntriesChangeRequired
	default:
		return fmt.Errorf("topic operation failed: %s", code)
	}
}

func mapTemplateErrorCode(code v1.ErrorCode) error {
	switch code {
	case v1.ErrorCode_ERROR_CODE_UNSPECIFIED:
		return nil
	case v1.ErrorCode_ERROR_CODE_BAD_DATA:
		return ErrBadRequest
	default:
		return fmt.Errorf("template operation failed: %s", code)
	}
}
