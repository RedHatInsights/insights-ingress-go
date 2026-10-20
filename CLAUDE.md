@AGENTS.md

## Build & Test Commands

- **Build**: `make build`
- **Test**: `make test` (runs `go test -p 1 -v ./...` with correct platform tags)
- **macOS prerequisite**: `brew install pkg-config librdkafka` (builds use `-tags dynamic` automatically)

## CI Checks to Match

- PR workflow runs `go test ./...` and validates `internal/api/openapi.json` with `openapi-spec-validator`
- If you change the OpenAPI spec, ensure it remains valid JSON and passes validation

## Key Gotchas

- Tests require sequential execution (`-p 1`); parallel packages will fail
- The `ACG_CONFIG` env var must point to `cdappconfig.json` for CI-mode tests
- `development/.env` contains local-only MinIO credentials committed intentionally; do not add real secrets
- The binary output (`insights-ingress-go`) is gitignored; do not commit it
