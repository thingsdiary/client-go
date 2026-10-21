package client

import (
	"context"

	"github.com/centrifugal/centrifuge-go"
)

type mockRealtimeConn struct {
	state centrifuge.State
	rpc   func(context.Context, string, []byte) (centrifuge.RPCResult, error)
}

func (c *mockRealtimeConn) State() centrifuge.State {
	return c.state
}

func (c *mockRealtimeConn) RPC(ctx context.Context, method string, data []byte) (centrifuge.RPCResult, error) {
	return c.rpc(ctx, method, data)
}

func (c *mockRealtimeConn) Close() {
	c.state = centrifuge.StateClosed
}
