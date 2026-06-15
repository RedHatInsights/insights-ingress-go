# AGENTS.md

## Project Overview

**insights-ingress-go** is a Go service in the Red Hat cloud.redhat.com platform that receives file uploads from clients, stages them to object storage (S3-compatible or filesystem), and announces them to downstream services via Kafka. It runs inside OpenShift behind a 3Scale authentication gateway.

- **Language**: Go (1.25+)
- **Module**: `github.com/redhatinsights/insights-ingress-go`
- **Entry point**: `cmd/insights-ingress/main.go`
- **All business logic**: `internal/` (no exported library packages)

## Quick Reference

| Task | Command |
|------|---------|
| Build | `make build` |
| Test (all) | `make test` |
| Test (CI mode) | `ACG_CONFIG="$(pwd)/cdappconfig.json" go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...` |
| Run locally (S3/MinIO) | `make start-api-dependencies && make run-api` |
| Run locally (filesystem) | `make start-filebased-api-dependencies && make run-filebased-api` |
| Test upload | `make run-upload-test` |
| Docker build | `docker build -f Dockerfile -t ingress:latest .` |

**macOS note**: On macOS, builds and tests require `-tags dynamic` (already set in the Makefile). You also need `brew install pkg-config librdkafka`.

## Cross-Cutting Conventions

### Configuration

- All config goes through `internal/config/config.go` using Viper with env prefix `INGRESS_`.
- Clowder integration is gated behind `clowder.IsClowderEnabled()`.
- See [docs/configuration-guidelines.md](docs/configuration-guidelines.md) for details.

### Logging

- Use the global `l.Log` logger (import aliased as `l`). Never use `fmt.Println` or standard `log`.
- Always use `logrus.Fields` for structured logging. Include `request_id` on every request-scoped log.
- See [docs/logging-and-observability-guidelines.md](docs/logging-and-observability-guidelines.md) for details.

### Error Handling

- HTTP handlers write status codes and return directly; they do not return errors.
- Use `logrus.Fields{"error": err}` for attaching errors. Use `Fatal` only in `main.go` for startup failures.
- See [docs/error-handling-guidelines.md](docs/error-handling-guidelines.md) for details.

### Testing

- Framework: Ginkgo v1 + Gomega (dot-imported). Do not use Ginkgo v2.
- Each test package has a `*_suite_test.go` bootstrap file that calls `l.InitLogger(config.Get())`.
- Use in-repo fakes (`stage.Fake`, `validators.Fake`, `announcers.Fake`) instead of mocking.
- Tests must run sequentially: `go test -p 1 -v ./...`
- See [docs/testing-guidelines.md](docs/testing-guidelines.md) for details.

### Interfaces and Dependency Injection

- Define interfaces in `types.go` within each domain package.
- Implementations live in sub-packages (e.g., `stage/s3compat/`, `stage/filebased/`, `validators/kafka/`).
- Fakes live in `fake.go` alongside the interface.
- Handlers are constructed via `NewHandler(...)` functions that accept interface dependencies.

### Metrics

- Use `prometheus/client_golang/prometheus/promauto` for auto-registration.
- Prefix all metric names with `ingress_`.
- Declare metrics in dedicated `metrics.go` files.
- Normalize user-agent strings via `NormalizeUserAgent()` before using as labels.

## Key Design Decisions

1. **Single Kafka announce topic**: All upload types go to `platform.upload.announce` with a `service` Kafka header for consumer filtering.
2. **Content-type routing**: Service name extracted from `application/vnd.redhat.<service>.<category>` MIME type pattern.
3. **Stager abstraction**: Storage backend is pluggable via the `stage.Stager` interface (S3 or filesystem), selected at runtime via `StagerImplementation` config.
4. **Produce-only Kafka architecture**: Two independent Kafka producers run in dedicated goroutines — one for validation announcements (`platform.upload.announce`), one for status tracking (`platform.payload-status`). The service does not consume messages.
5. **Retry by requeue**: Failed Kafka publishes are re-enqueued to the channel with no backoff.

## CI/CD

- **GitHub Actions** (`pr.yml`): Runs `go test ./...` and validates `internal/api/openapi.json` on PRs.
- **Tekton/Konflux** (`.tekton/`): Hermetic builds with Go module prefetch on `master` branch.
- **App-SRE** (`pr_check.sh`, `build_deploy.sh`): Bonfire-based ephemeral environment testing and quay.io image publishing.
- **Renovate** (`renovate.json`): Automated Go dependency updates.
- **Dependabot** (`.github/dependabot.yml`): Docker base image updates on `security-compliance` branch only.

## Detailed Guidelines Index

Each file below contains specific, enforceable rules for its domain. Read the relevant file before making changes in that area.

| Domain | File |
|--------|------|
| API contracts and endpoints | [docs/api-contracts-guidelines.md](docs/api-contracts-guidelines.md) |
| Kafka messaging and producers | [docs/async-and-messaging-guidelines.md](docs/async-and-messaging-guidelines.md) |
| Package structure and patterns | [docs/code-organization-guidelines.md](docs/code-organization-guidelines.md) |
| Viper/Clowder configuration | [docs/configuration-guidelines.md](docs/configuration-guidelines.md) |
| Content-type parsing and routing | [docs/content-type-routing-guidelines.md](docs/content-type-routing-guidelines.md) |
| Input validation and metadata | [docs/data-validation-guidelines.md](docs/data-validation-guidelines.md) |
| Go modules and dependencies | [docs/dependency-management-guidelines.md](docs/dependency-management-guidelines.md) |
| Dockerfiles, ClowdApp, CI/CD | [docs/deployment-guidelines.md](docs/deployment-guidelines.md) |
| Error codes and error propagation | [docs/error-handling-guidelines.md](docs/error-handling-guidelines.md) |
| Storage, Kafka, and platform integration | [docs/integration-guidelines.md](docs/integration-guidelines.md) |
| Logging, metrics, and tracing | [docs/logging-and-observability-guidelines.md](docs/logging-and-observability-guidelines.md) |
| Memory limits, concurrency, profiling | [docs/performance-guidelines.md](docs/performance-guidelines.md) |
| Auth, TLS, deny lists, input sanitization | [docs/security-guidelines.md](docs/security-guidelines.md) |
| S3 and filesystem staging patterns | [docs/storage-patterns-guidelines.md](docs/storage-patterns-guidelines.md) |
| Ginkgo/Gomega testing conventions | [docs/testing-guidelines.md](docs/testing-guidelines.md) |
