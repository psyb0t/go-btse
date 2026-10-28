# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.4, 2026-10-28

- Updated Semgrep in the development image to 1.182.1 and refreshed its locked dependencies. certifi 2026.10.23 drops the emSign root CAs whose intermediate keys were compromised (GHSA-6q48-6c7w-9gw2).
- Dependabot no longer opens pull requests for `mcp` or the OpenTelemetry packages in `requirements-semgrep.txt`. Semgrep pins their versions itself, so bumping one of them on its own cannot install.
- Updated kin-openapi, which the code generator uses, from 0.142.0 to 0.149.0. The newer version fixes several panics on malformed specs reported in GitHub security advisories.
- The security scan merges its two gitleaks runs into one SARIF run. GitHub code scanning rejects a report with two runs from the same tool, so no scan results had been uploaded.

## v0.1.3, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.
- Raised the minimum Go version in `go.mod` from 1.26.9 to 1.26.10, which fixes the `crypto/x509` vulnerability GO-2026-6809 found by govulncheck.
- The issue-pull workflow runs at hours derived from the repository name instead of the same four hours as every other repository.

## v0.1.2, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.

## v0.1.1, 2026-10-17

- Generated doc comments keep BTSE's own description text. The REST clients now generate from the downloaded specs unchanged, and the WebSocket converter no longer rewrites dashes in descriptions. The `patch-spec` tool and `internal/specprose` are removed.
- `make sec` uses Semgrep 1.181.0. Its pinned requirements now hash every published file, so the development image also builds on arm64.

## v0.1.0, 2026-10-17

- Added REST clients generated with oapi-codegen from BTSE's OpenAPI 3.0.3 documents at `https://docs.btse.com/specs/{product}-rest.openapi.yaml`: spot and earn spec version 3.3.0, futures 2.3.0, and markets, wallet and OTC 1.0.0. Each complete spec has its own package.
- Added spot, futures and OTC WebSocket clients generated with gen-asyncapi from `https://docs.btse.com/specs/{product}-websocket.asyncapi.yaml`: spot and futures spec version 3.2.0 and OTC 1.0.0. Generation converts AsyncAPI 2.6 to 3.0 and repairs the provider's payload schemas and ambiguous message envelopes.
- Generation normalizes dash punctuation in provider descriptions, summaries and titles for REST and WebSocket comments. Schema values, operation IDs, examples and wire data stay unchanged.
- Added a tested HMAC-SHA384 request signer and local HTTP tests. Number fields use `float64`.
- Added vendored dependencies, development tooling, regeneration workflows and the `go-btse` agent skill. Downloaded specifications stay gitignored.
