package client

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/thingsdiary/client-go/openapi"
)

func restoreResponseError(resp *http.Response, notFoundError error) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return notFoundError
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusGone:
		return ErrEntityRemoved
	case http.StatusBadRequest:
		var errorResp openapi.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil &&
			errorResp.ErrorCode == openapi.ResponseErrorCodeEntityNotDeleted {
			return ErrEntityNotDeleted
		}

		return errors.Wrap(ErrBadRequest, resp.Status)
	default:
		return errors.Errorf("unexpected status code: %s", resp.Status)
	}
}
