package btsews_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	btsews "github.com/psyb0t/go-btse/ws"
	futuresws "github.com/psyb0t/go-btse/ws/futures"
	otcws "github.com/psyb0t/go-btse/ws/otc"
	spotsws "github.com/psyb0t/go-btse/ws/spot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const snapshotMessageName = "receiveOrderbookUpdateOrderbookSnapshot"

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	return data
}

func TestDecodeSpotFrames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		file string
		name string
	}{
		{"spot_l1.json", "receiveL1SnapshotreceiveL1Snapshot"},
		{"spot_snapshot.json", snapshotMessageName},
		{"spot_ungrouped.json", snapshotMessageName},
		{"spot_delta.json", "receiveOrderbookUpdateOrderbookDelta"},
		{"spot_trade.json", "receivePublicTradesreceivePublicTrades"},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()
			message, err := spotsws.DecodeReceivedMessage(fixture(t, tc.file))
			require.NoError(t, err)
			assert.Equal(t, tc.name, message.Name)
		})
	}
}

func TestDecodeFuturesFrames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		file string
		name string
	}{
		{"futures_trade.json", "receivePublicTradesreceivePublicTrades"},
		{"futures_ungrouped.json", snapshotMessageName},
		{"futures_positions.json", "receivePositionsreceivePositions"},
		{
			"futures_all_positions.json",
			"receiveAllPositionsreceiveAllPositions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()
			message, err := futuresws.DecodeReceivedMessage(fixture(t, tc.file))
			require.NoError(t, err)
			assert.Equal(t, tc.name, message.Name)
		})
	}
}

func TestDecodeOTCQuote(t *testing.T) {
	t.Parallel()

	message, err := otcws.DecodeReceivedMessage(fixture(t, "otc_quote.json"))
	require.NoError(t, err)
	assert.Equal(t, "receiveQuoteQuoteStreamMessage", message.Name)
}

func TestUnknownFrames(t *testing.T) {
	t.Parallel()

	_, err := spotsws.DecodeReceivedMessage([]byte(`{"topic":"unknown"}`))
	require.ErrorIs(t, err, asyncapirt.ErrUnknownMessage)
	_, err = futuresws.DecodeReceivedMessage([]byte(`{`))
	require.ErrorIs(t, err, commerr.ErrCouldNotDecode)
}

type recordingTransport struct {
	written []byte
	read    []byte
}

func (t *recordingTransport) WriteMessage(
	_ context.Context,
	data []byte,
) error {
	t.written = data

	return nil
}

func (t *recordingTransport) ReadMessage(context.Context) ([]byte, error) {
	return t.read, nil
}

func (*recordingTransport) Close() error {
	return nil
}

func TestHeartbeatTransport(t *testing.T) {
	t.Parallel()

	raw := &recordingTransport{read: []byte("pong")}
	transport := btsews.HeartbeatTransport{Transport: raw}
	client := spotsws.NewClient(transport)
	require.NoError(t, client.SendPing(context.Background(), "ping"))
	assert.Equal(t, "ping", string(raw.written))

	frame, err := transport.ReadMessage(context.Background())
	require.NoError(t, err)
	message, err := spotsws.DecodeReceivedMessage(frame)
	require.NoError(t, err)
	assert.Equal(t, "receivePongreceivePong", message.Name)
}
