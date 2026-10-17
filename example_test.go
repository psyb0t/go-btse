package btse_test

import (
	"context"
	"os"

	"github.com/psyb0t/go-btse"
	btseauth "github.com/psyb0t/go-btse/auth"
)

// The README example requires testnet credentials and only compiles.
//
//nolint:testableexamples // Requires real testnet credentials.
func Example() {
	signer, err := btseauth.NewSigner(
		os.Getenv("BTSE_API_KEY"), os.Getenv("BTSE_API_SECRET"),
	)
	if err != nil {
		panic(err)
	}

	client, err := btse.NewClientWithResponses(
		"https://testapi.btse.io/spot",
		btse.WithRequestEditorFn(signer.RequestEditor),
	)
	if err != nil {
		panic(err)
	}

	response, err := client.GetAccountFeesWithResponse(
		context.Background(), &btse.GetAccountFeesParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = response.JSON200
}
