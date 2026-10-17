// Package btsews provides transport helpers for BTSE's WebSocket clients.
package btsews

import (
	"bytes"
	"context"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
)

const (
	ping     = "ping"
	pong     = "pong"
	jsonPing = `"ping"`
	jsonPong = `"pong"`
)

// HeartbeatTransport adapts generated JSON heartbeats to BTSE's plain text.
// Wrap the spot or futures transport before passing it to NewClient.
type HeartbeatTransport struct {
	asyncapirt.Transport
}

// WriteMessage emits ping as plain text, as BTSE requires.
func (t HeartbeatTransport) WriteMessage(
	ctx context.Context,
	data []byte,
) error {
	if bytes.Equal(data, []byte(jsonPing)) {
		data = []byte(ping)
	}

	if err := t.Transport.WriteMessage(ctx, data); err != nil {
		return ctxerrors.Wrap(err, "write BTSE frame")
	}

	return nil
}

// ReadMessage turns plain text pong into the generated decoder's JSON string.
func (t HeartbeatTransport) ReadMessage(ctx context.Context) ([]byte, error) {
	data, err := t.Transport.ReadMessage(ctx)
	if err != nil {
		return nil, ctxerrors.Wrap(err, "read BTSE frame")
	}

	if bytes.Equal(data, []byte(pong)) {
		return []byte(jsonPong), nil
	}

	return data, nil
}
