package client

import (
	"context"
	"testing"

	"github.com/centrifugal/centrifuge-go"
	"github.com/samber/mo"
	"github.com/stretchr/testify/require"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

func TestClientDeleteEntityHardDeleteRPC(t *testing.T) {
	tests := []struct {
		name   string
		method string
		call   func(client *Client) error
		check  func(t *testing.T, payload []byte)
	}{
		{
			name:   "entry",
			method: "io.thingsdiary.v1.entry.delete",
			call: func(client *Client) error {
				return client.DeleteEntry(
					context.Background(),
					"diary-1",
					"entry-1",
					DeleteEntryParams{Soft: mo.Some(false)},
				)
			},
			check: func(t *testing.T, payload []byte) {
				t.Helper()

				var request v1.DeleteEntryRequest
				require.NoError(t, proto.Unmarshal(payload, &request))
				require.NotNil(t, request.Soft)
				require.False(t, request.GetSoft())
			},
		},
		{
			name:   "topic",
			method: "io.thingsdiary.v1.topic.delete",
			call: func(client *Client) error {
				return client.DeleteTopic(
					context.Background(),
					"diary-1",
					"topic-1",
					DeleteTopicParams{Soft: mo.Some(false)},
				)
			},
			check: func(t *testing.T, payload []byte) {
				t.Helper()

				var request v1.DeleteTopicRequest
				require.NoError(t, proto.Unmarshal(payload, &request))
				require.NotNil(t, request.Soft)
				require.False(t, request.GetSoft())
			},
		},
		{
			name:   "template",
			method: "io.thingsdiary.v1.template.delete",
			call: func(client *Client) error {
				return client.DeleteTemplate(
					context.Background(),
					"diary-1",
					"template-1",
					DeleteTemplateParams{Soft: mo.Some(false)},
				)
			},
			check: func(t *testing.T, payload []byte) {
				t.Helper()

				var request v1.DeleteTemplateRequest
				require.NoError(t, proto.Unmarshal(payload, &request))
				require.NotNil(t, request.Soft)
				require.False(t, request.GetSoft())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient()
			client.credentials = &Credentials{SigningPrivateKey: testSigningPrivateKey()}
			client.conn = &mockRealtimeConn{
				state: centrifuge.StateConnected,
				rpc: func(_ context.Context, method string, data []byte) (centrifuge.RPCResult, error) {
					require.Equal(t, test.method, method)
					test.check(t, rpcPayload(t, method, data))

					return centrifuge.RPCResult{}, nil
				},
			}

			err := test.call(client)

			require.NoError(t, err)
		})
	}
}
