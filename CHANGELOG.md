# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.1, 2026-10-17

- Generated doc comments keep BTSE's own description text. The REST clients now generate from the downloaded specs unchanged, and the WebSocket converter no longer rewrites dashes in descriptions. The `patch-spec` tool and `internal/specprose` are removed.
- `make sec` uses Semgrep 1.181.0. Its pinned requirements now hash every published file, so the development image also builds on arm64.

## v0.1.0, 2026-10-17

- Added REST clients generated with oapi-codegen from BTSE's OpenAPI 3.0.3 documents at `https://docs.btse.com/specs/{product}-rest.openapi.yaml`: spot and earn spec version 3.3.0, futures 2.3.0, and markets, wallet and OTC 1.0.0. Each complete spec has its own package.
- Added spot, futures and OTC WebSocket clients generated with gen-asyncapi from `https://docs.btse.com/specs/{product}-websocket.asyncapi.yaml`: spot and futures spec version 3.2.0 and OTC 1.0.0. Generation converts AsyncAPI 2.6 to 3.0 and repairs the provider's payload schemas and ambiguous message envelopes.
- Generation normalizes dash punctuation in provider descriptions, summaries and titles for REST and WebSocket comments. Schema values, operation IDs, examples and wire data stay unchanged.
- Added a tested HMAC-SHA384 request signer and local HTTP tests. Number fields use `float64`.
- Added vendored dependencies, development tooling, regeneration workflows and the `go-btse` agent skill. Downloaded specifications stay gitignored.
