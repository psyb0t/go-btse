---
name: go-btse
description: Unofficial Go client for BTSE, generated from its public OpenAPI and AsyncAPI specifications. REST packages cover spot and futures orders, account fees, positions, public markets, wallets, transfers, withdrawals, earn products and OTC quotes. WebSocket packages cover spot, futures and OTC messages. Use when Go code needs BTSE API access or already imports github.com/psyb0t/go-btse.
homepage: https://github.com/psyb0t/go-btse
user-invocable: true
metadata:
  openclaw:
    requires:
      bins:
        - go
permissions:
  filesystem: "Guides edits to Go source in the current project. The client reads and writes no files."
  shell: "Uses the Go toolchain to add and compile the dependency."
  network: "Module installation fetches dependencies from the Go module proxy. Runtime requests send API parameters and authentication headers to the selected BTSE HTTPS or WSS host. Tests use local servers and fixtures."
---

# go-btse

Unofficial typed Go client for BTSE. The root `btse` package covers spot REST operations. Other products have separate packages to keep their schemas and operation names distinct.

See [references/setup.md](references/setup.md) for installation, product base URLs, responses and generation details.

## Security & safety

- Read private API credentials from `BTSE_API_KEY` and `BTSE_API_SECRET`. Never hardcode, commit or log them, or print signed request headers. Public market data does not require credentials.
- Default to `https://testapi.btse.io`. Production requests use `https://api.btse.com`. Select the service prefix from the table in the setup reference, and use credentials for that environment.
- Orders, cancellations, position changes, transfers, withdrawals, earn deposits and redemptions, and accepted OTC quotes change the account. Call them only for the exact action the user requested. Use production only when requested.

## When to use

- Go code needs BTSE account data, public market data or WebSocket messages.
- Go code needs to place orders or use BTSE wallet, earn or OTC operations at the user's request.
- The project already imports `github.com/psyb0t/go-btse` or one of its product packages.

## When NOT to use

- The code is not Go.
- The user needs an SDK maintained by BTSE. This client is unofficial.
- The task changes the generated client. Read the repository's `AGENTS.md` for that workflow.

## Usage

Construct a signer with `btseauth.NewSigner(key, secret)`. Handle its error, then pass `btse.WithRequestEditorFn(signer.RequestEditor)` to `btse.NewClientWithResponses("https://testapi.btse.io/spot", opts...)`. For another product, use that package's constructor and `WithRequestEditorFn` option with its base URL.

For example, `client.GetAccountFeesWithResponse(ctx, &btse.GetAccountFeesParams{})` fetches spot maker and taker fees. The complete compiling example is `example_test.go`, also shown in the README.

The signer sets `request-api`, an increasing millisecond `request-nonce`, and the lowercase hexadecimal HMAC-SHA384 `request-sign`. Register it after editors that change the URL or body. Share one signer per API key. It signs the API path without the query, strips `/spot` or `/futures` only before `/api/`, and preserves the exact body bytes.

The Go error covers transport and decoding failures. Check `StatusCode()` and the expected response field before dereferencing it. A non-2xx API result can return a nil Go error. Most generated fields are pointers. JSON number fields use `float64`; fields declared as strings stay strings.

Method names come from each product's operation IDs. `CreateOrderWithResponse` exists in both the root spot package and the `futures` package, with different request types. Keep those types in their owning package.
