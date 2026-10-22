package client

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/centrifugal/centrifuge-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	v1 "github.com/thingsdiary/client-go/proto"
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

func testSigningPrivateKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(make([]byte, ed25519.SeedSize))
}

func rpcPayload(t *testing.T, method string, data []byte) []byte {
	t.Helper()

	var signed v1.SignedEnvelope
	require.NoError(t, proto.Unmarshal(data, &signed))
	require.True(t, ed25519.Verify(testSigningPrivateKey().Public().(ed25519.PublicKey), signed.GetEnvelope(), signed.GetSignature()))

	var envelope v1.Envelope
	require.NoError(t, proto.Unmarshal(signed.GetEnvelope(), &envelope))
	require.Equal(t, envelopeVersion, envelope.GetVersion())
	require.Equal(t, method, envelope.GetMethod())
	require.NoError(t, uuid.Validate(envelope.GetOperationId()))

	return envelope.GetPayload()
}
