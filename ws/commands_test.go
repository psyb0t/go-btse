package btsews_test

import (
	"context"
	"testing"

	spotsws "github.com/psyb0t/go-btse/ws/spot"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type snapshotPayload = spotsws.ReceiveL1SnapshotreceiveL1SnapshotPayload

func TestOrderbookCommandsHaveTypedPayloads(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{}
	client := spotsws.NewClient(transport)
	request := spotsws.SubscriptionRequest{
		Op:   spotsws.SubscriptionRequestOpSubscribe,
		Args: []string{"snapshotL1:BTC-USD"},
	}
	require.NoError(t, client.SubscribeL1SnapshotL1Subscribe(
		context.Background(), request))
	assert.JSONEq(t,
		`{"op":"subscribe","args":["snapshotL1:BTC-USD"]}`,
		string(transport.written))

	request.Args = []string{"update:BTC-USD_0"}
	require.NoError(t, client.SubscribeOrderbookUpdateL2Subscribe(
		context.Background(), request))
	assert.JSONEq(t,
		`{"op":"subscribe","args":["update:BTC-USD_0"]}`,
		string(transport.written))
}

func TestSnapshotPreservesPriceStrings(t *testing.T) {
	t.Parallel()

	message, err := spotsws.DecodeReceivedMessage(fixture(t, "spot_l1.json"))
	require.NoError(t, err)

	payload, ok := message.Payload.(snapshotPayload)
	require.True(t, ok)
	require.NotNil(t, payload.Data)
	assert.Equal(t, [][]string{{"12.5", "1.25"}}, payload.Data.Bids)
	assert.Equal(t, [][]string{{"12.6", "2.0"}}, payload.Data.Asks)
}
