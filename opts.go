package client

import (
	"net/url"
	"time"
)

type options struct {
	baseURL     string
	realtimeURL string
	timeout     time.Duration
}

func defaultOptions() *options {
	return &options{
		baseURL: "https://cloud.thingsdiary.io/api",
		timeout: 5 * time.Second,
	}
}

func realtimeURLFromBaseURL(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		return ""
	}

	u.Path = "/connection/websocket"
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

type clientOption func(o *options)

func WithBaseURL(baseURL string) clientOption {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

// WithRealtimeURL configures the Centrifugo WebSocket endpoint.
func WithRealtimeURL(url string) clientOption {
	return func(o *options) { o.realtimeURL = url }
}

func WithTimeout(timeout time.Duration) clientOption {
	return func(o *options) {
		if timeout >= 0 {
			o.timeout = timeout
		}
	}
}
