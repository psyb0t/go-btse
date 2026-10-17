package btse_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-btse"
	btseauth "github.com/psyb0t/go-btse/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRESTCalls(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		path string
		body string
	}{
		{
			"fees", "/spot/api/v3.3/user/fees",
			`[{"symbol":"SYMBOL_X","makerFee":0.002}]`,
		},
		{
			"orders", "/spot/api/v3.3/order",
			`{"orderID":"order-x","price":123.456789}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, tc.path, r.URL.Path)
					assert.Equal(t, "test-key", r.Header.Get("request-api"))
					assert.NotEmpty(t, r.Header.Get("request-sign"))
					assert.NotEmpty(t, r.Header.Get("request-nonce"))
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(tc.body))
				},
			))
			t.Cleanup(server.Close)

			signer, err := btseauth.NewSigner("test-key", "test-secret")
			require.NoError(t, err)

			client, err := btse.NewClientWithResponses(
				server.URL+"/spot",
				btse.WithRequestEditorFn(signer.RequestEditor),
			)
			require.NoError(t, err)

			if tc.name == "fees" {
				response, callErr := client.GetAccountFeesWithResponse(
					context.Background(), &btse.GetAccountFeesParams{},
				)
				require.NoError(t, callErr)
				require.NotNil(t, response.JSON200)
				require.Len(t, *response.JSON200, 1)
				assert.InDelta(t, 0.002,
					*(*response.JSON200)[0].MakerFee, 0)

				return
			}

			response, err := client.GetOrderWithResponse(
				context.Background(), &btse.GetOrderParams{},
			)
			require.NoError(t, err)
			require.NotNil(t, response.JSON200)
			assert.Equal(t, "order-x", *response.JSON200.OrderID)
			assert.InDelta(t, 123.456789, *response.JSON200.Price, 0)
		})
	}
}
