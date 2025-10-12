package client

import "time"

type options struct {
	baseURL string
	timeout time.Duration
}

func defaultOptions() *options {
	return &options{
		baseURL: "https://cloud.thingsdiary.io",
		timeout: 5 * time.Second,
	}
}

type clientOption func(o *options)

func WithBaseURL(baseURL string) clientOption {
	return func(o *options) {
		o.baseURL = baseURL
	}
}

func WithTimeout(timeout time.Duration) clientOption {
	return func(o *options) {
		if timeout >= 0 {
			o.timeout = timeout
		}
	}
}
