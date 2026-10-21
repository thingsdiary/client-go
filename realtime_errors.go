package client

import (
	"errors"
	"fmt"
	"strings"

	"github.com/centrifugal/centrifuge-go"
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
