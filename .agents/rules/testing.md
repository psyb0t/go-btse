# Testing

Refines `~/.claude/rule-details/go/writing-tests.md`.

- **TST1** Tests use the owning package's external `_test` package and its exported API. Command tests may use package `main` for the conversion boundary.
- **TST2** Tests never call BTSE. They run the client against an `httptest` server.
- **TST3** The README usage example lives in `example_test.go` too, so `make test` compiles it. A README example change goes into `example_test.go` in the same commit.
- **TST4** Every spec patch or config change that alters wire behavior gets a test with a body shaped like what BTSE sends.
- **TST5** Table tests use `testCases` and `tc`, call `t.Parallel()` in the test and in every subtest, and assert errors with `require.ErrorIs` against a sentinel.
- **TST6** `make test` runs `go test -race` and must pass.
