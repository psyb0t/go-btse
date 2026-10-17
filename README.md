# go-btse

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-btse.svg)](https://pkg.go.dev/github.com/psyb0t/go-btse)
[![CI](https://github.com/psyb0t/go-btse/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-btse/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-btse/badges/version.svg)](https://github.com/psyb0t/go-btse/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-btse/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-btse/badges/importers.svg)](https://github.com/psyb0t/go-btse/blob/badges/importers.md)

Unofficial Go client for the [BTSE API](https://docs.btse.com), generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from BTSE's public OpenAPI documents. I am not affiliated with BTSE.

The root package is `btse` for spot trading. Separate `futures`, `markets`, `wallet`, `earn` and `otc` packages cover the other REST products. WebSocket packages live under `ws/spot`, `ws/futures` and `ws/otc`. Generated files come from the full specifications. Regenerate them instead of editing them.

## Install

```sh
go get github.com/psyb0t/go-btse
```

## Usage

Every REST endpoint in the specs has a typed method on its package's `ClientWithResponses`. Responses decode into `JSON200` and the other status fields.

```go
package main

import (
	"context"
	"os"

	"github.com/psyb0t/go-btse"
	btseauth "github.com/psyb0t/go-btse/auth"
)

func main() {
	signer, err := btseauth.NewSigner(
		os.Getenv("BTSE_API_KEY"),
		os.Getenv("BTSE_API_SECRET"),
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
```

The same code is in `example_test.go`, so `make test` checks that it compiles. Check `StatusCode()` and the decoded response field before using a result. An API rejection can return a nil Go error.

## Auth

BTSE private endpoints require `request-api` (API key), `request-nonce` (milliseconds) and `request-sign` (hex HMAC-SHA384). The signature input is the API path, nonce and exact body bytes concatenated, without the query string. `btseauth.NewSigner` supplies these headers through the generated `RequestEditorFn` interface, as shown above. The signer strips the `/spot` or `/futures` service prefix before `/api/` when signing. Public market data works without a signer.

Share one signer per key to allocate increasing nonces, and register its request editor after editors that change the URL or body. Use credentials for the selected environment and the permissions required by the operation. Never log authentication headers.

| Packages | Production base URL | Testnet base URL |
|---|---|---|
| `btse`, `earn` | `https://api.btse.com/spot` | `https://testapi.btse.io/spot` |
| `futures` | `https://api.btse.com/futures` | `https://testapi.btse.io/futures` |
| `markets`, `wallet`, `otc` | `https://api.btse.com` | `https://testapi.btse.io` |

Each package has its own `WithRequestEditorFn` option. Pass `signer.RequestEditor` to that option when constructing its client.

## Regenerating

The specs are not committed. `make generate` downloads the current documents into the gitignored `api/` directory and regenerates every package:

```sh
make generate
```

Each output package has a `gen.go` that declares its download and generation steps. REST specs come from `https://docs.btse.com/specs/{product}-rest.openapi.yaml`, for `spot`, `futures`, `markets`, `wallet`, `earn` and `otc`. `scripts/patch-spec.sh` runs `internal/cmd/patch-spec` to write `api/{product}-rest.openapi.fixed.yaml`. It normalizes dash punctuation in descriptions, summaries and titles, preserving schema values, operation IDs, examples and wire data. `config.yaml` maps `number` to `float64` in every REST package.

WebSocket specs come from `https://docs.btse.com/specs/{spot,futures,otc}-websocket.asyncapi.yaml`. Their AsyncAPI 2.6 documents are converted to AsyncAPI 3.0 before the pinned `go tool gen-asyncapi` generates the WebSocket packages. The converter repairs misplaced payload schemas and example-only subscription payloads, separates received messages by channel topic and applies the same prose normalization through `internal/specprose`. Spot and futures transports use `ws.HeartbeatTransport` to adapt generated JSON heartbeats to BTSE's plain text `ping` and `pong` frames.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-btse) teaches a coding agent how to use go-btse when it writes Go. It works in any agent that reads `.agents/skills/`.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-btse@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-btse@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo. Marketplace installs require the go-btse entry in the central catalog.

### OpenClaw

The release workflow publishes the skill to ClawHub on tag pushes. After publication:

```bash
openclaw skills install @psyb0t/go-btse
```

## Development

Every command runs in the `Dockerfile.dev` image. The build host needs Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the specs and regenerate
```

## License

[MIT](LICENSE). The generated code comes from BTSE's published API specifications. The specifications have no explicit license field and are not redistributed here. The repository license does not grant rights to BTSE's API or its documentation.
