package client

import (
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	localRealtimePort = "8000"
	realtimeAPIPath   = "/v1/realtime"
	rawRealtimePath   = "/connection/websocket"
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
	if err != nil || u.Hostname() == "" {
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

	if isLocalhost(u.Hostname()) {
		u.Scheme = "ws"
		u.Host = net.JoinHostPort(u.Hostname(), localRealtimePort)
		u.Path = rawRealtimePath
	} else {
		u.Path = strings.TrimRight(u.Path, "/") + realtimeAPIPath
	}
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

func isLocalhost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
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
