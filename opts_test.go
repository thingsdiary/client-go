package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRealtimeURLFromBaseURL(t *testing.T) {
	tests := map[string]struct {
		baseURL string
		want    string
	}{
		"production": {
			baseURL: "https://cloud.thingsdiary.io/api",
			want:    "wss://cloud.thingsdiary.io/api/v1/realtime",
		},
		"remote HTTP with trailing slash": {
			baseURL: "http://example.test:8081/api/",
			want:    "ws://example.test:8081/api/v1/realtime",
		},
		"localhost": {
			baseURL: "http://localhost:8081/api",
			want:    "ws://localhost:8000/connection/websocket",
		},
		"localhost always uses the plain local socket": {
			baseURL: "https://localhost/api",
			want:    "ws://localhost:8000/connection/websocket",
		},
		"IPv4 loopback": {
			baseURL: "http://127.0.0.1:8081/api",
			want:    "ws://127.0.0.1:8000/connection/websocket",
		},
		"IPv6 loopback": {
			baseURL: "http://[::1]:8081/api",
			want:    "ws://[::1]:8000/connection/websocket",
		},
		"query and fragment are removed": {
			baseURL: "https://example.test/api?debug=true#fragment",
			want:    "wss://example.test/api/v1/realtime",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.want, realtimeURLFromBaseURL(test.baseURL))
		})
	}
}

func TestRealtimeURLFromBaseURLRejectsInvalidBaseURL(t *testing.T) {
	require.Empty(t, realtimeURLFromBaseURL("not a URL"))
	require.Empty(t, realtimeURLFromBaseURL("ftp://example.test/api"))
}
