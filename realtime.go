package client

import (
	"bytes"
	"context"
	"crypto/ed25519"

	"github.com/centrifugal/centrifuge-go"
	"github.com/pkg/errors"
	v1 "github.com/thingsdiary/client-go/proto"
	"google.golang.org/protobuf/proto"
)

type realtimeConn interface {
	State() centrifuge.State
	RPC(context.Context, string, []byte) (centrifuge.RPCResult, error)
	Close()
}

// Connect obtains a connection JWT and waits until the realtime connection is ready.
func (c *Client) Connect(ctx context.Context) error {
	if c.authToken == "" {
		return ErrUnauthorized
	}

	c.realtimeMu.Lock()

	if c.conn != nil {
		if c.conn.State() == centrifuge.StateConnected {
			c.realtimeMu.Unlock()
			return nil
		}

		c.realtimeMu.Unlock()
		return errors.New("realtime client is already connecting")
	}

	if c.realtimeURL == "" {
		c.realtimeMu.Unlock()
		return errors.New("realtime URL is not configured")
	}

	token, err := c.GetRealtimeToken(ctx)
	if err != nil {
		c.realtimeMu.Unlock()
		return err
	}

	events := make(chan *v1.Event, 64)
	var conn *centrifuge.Client

	conn = centrifuge.NewProtobufClient(c.realtimeURL, centrifuge.Config{
		Token: token,
		GetToken: func(centrifuge.ConnectionTokenEvent) (string, error) {
			return c.GetRealtimeToken(context.Background())
		},
	})

	connected := make(chan struct{}, 1)

	conn.OnConnected(func(centrifuge.ConnectedEvent) {
		select {
		case connected <- struct{}{}:
		default:
		}
	})

	conn.OnPublication(func(e centrifuge.ServerPublicationEvent) {
		var event v1.Event

		if err := proto.Unmarshal(e.Data, &event); err != nil {
			return
		}

		c.realtimeMu.RLock()
		defer c.realtimeMu.RUnlock()

		if c.conn != conn {
			return
		}

		select {
		case events <- &event:
		default:
		}
	})

	c.conn = conn
	c.events = events

	if err := conn.Connect(); err != nil {
		c.conn = nil
		c.events = nil
		close(events)
		c.realtimeMu.Unlock()
		conn.Close()
		return errors.Wrap(err, "failed to start realtime connection")
	}

	select {
	case <-connected:
		c.realtimeMu.Unlock()
		return nil
	case <-ctx.Done():
		c.conn = nil
		c.events = nil
		close(events)
		c.realtimeMu.Unlock()
		conn.Close()
		return ctx.Err()
	}
}

// Close closes the realtime connection and event stream. It is safe to call repeatedly.
func (c *Client) Close() {
	c.realtimeMu.Lock()
	conn := c.conn
	events := c.events
	c.conn = nil
	c.events = nil

	if events != nil {
		close(events)
	}

	c.realtimeMu.Unlock()

	if conn != nil {
		conn.Close()
	}
}

// Events returns publications decoded as typed protobuf events.
func (c *Client) Events() <-chan *v1.Event {
	c.realtimeMu.RLock()
	defer c.realtimeMu.RUnlock()

	return c.events
}

// rpc performs one named protobuf RPC. Mutations are deliberately never retried.
func (c *Client) rpc(ctx context.Context, method string, request, response proto.Message) error {
	c.realtimeMu.RLock()
	conn := c.conn
	c.realtimeMu.RUnlock()

	if conn == nil || conn.State() != centrifuge.StateConnected {
		return errors.New("realtime client is not connected")
	}

	data, err := proto.Marshal(request)
	if err != nil {
		return errors.Wrap(err, "failed to encode RPC request")
	}

	if c.credentials == nil {
		return ErrUnauthorized
	}

	envelope := NewEnvelope(method, data)
	signedEnvelope, err := signEnvelope(envelope, ed25519.PrivateKey(c.credentials.SigningPrivateKey))
	if err != nil {
		return err
	}

	data, err = proto.Marshal(signedEnvelope)
	if err != nil {
		return errors.Wrap(err, "failed to encode signed RPC envelope")
	}

	result, err := conn.RPC(ctx, method, data)
	if err != nil {
		return mapRealtimeRPCError(method, err)
	}

	if response == nil {
		return nil
	}

	if err := proto.Unmarshal(bytes.Clone(result.Data), response); err != nil {
		return errors.Wrap(err, "failed to decode RPC response")
	}

	return nil
}
