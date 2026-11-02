package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_DeleteEntity_HardDeleteQuery(t *testing.T) {
	hardDelete := false

	tests := []struct {
		name     string
		wantPath string
		call     func(client *Client) error
	}{
		{
			name:     "entry",
			wantPath: "/v1/diaries/diary-1/entries/entry-1",
			call: func(client *Client) error {
				return client.DeleteEntry(
					context.Background(),
					"diary-1",
					"entry-1",
					DeleteEntryParams{Soft: &hardDelete},
				)
			},
		},
		{
			name:     "topic",
			wantPath: "/v1/diaries/diary-1/topics/topic-1",
			call: func(client *Client) error {
				return client.DeleteTopic(
					context.Background(),
					"diary-1",
					"topic-1",
					DeleteTopicParams{Soft: &hardDelete},
				)
			},
		},
		{
			name:     "template",
			wantPath: "/v1/diaries/diary-1/templates/template-1",
			call: func(client *Client) error {
				return client.DeleteTemplate(
					context.Background(),
					"diary-1",
					"template-1",
					DeleteTemplateParams{Soft: &hardDelete},
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, test.wantPath, r.URL.Path)
				require.Equal(t, "false", r.URL.Query().Get("soft"))
				w.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(server.Close)

			client := NewClient(WithBaseURL(server.URL))
			client.credentials = &Credentials{}
			client.authToken = "token"

			// act
			err := test.call(client)

			// assert
			require.NoError(t, err)
		})
	}
}
