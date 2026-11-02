package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestoreResponseError_MapsAPIStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantError  error
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			wantError:  ErrEntryNotFound,
		},
		{
			name:       "permanently removed",
			statusCode: http.StatusGone,
			wantError:  ErrEntityRemoved,
		},
		{
			name:       "not deleted",
			statusCode: http.StatusBadRequest,
			body:       `{"error_code":"ENTITY_NOT_DELETED"}`,
			wantError:  ErrEntityNotDeleted,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// arrange
			resp := &http.Response{
				Status:     http.StatusText(test.statusCode),
				StatusCode: test.statusCode,
				Body:       http.NoBody,
			}
			if test.body != "" {
				resp.Body = io.NopCloser(strings.NewReader(test.body))
			}

			// act
			err := restoreResponseError(resp, ErrEntryNotFound)

			// assert
			if test.wantError == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, test.wantError)
		})
	}
}
