# Code Organization

## Package Layout

- Place the service entry point in `cmd/insights-ingress/main.go`; additional CLI tools go under `cmd/<tool-name>/`.
- Place all business logic under `internal/`; this codebase has no exported library packages.
- Organize `internal/` by domain concern, one directory per concern: `upload`, `stage`, `validators`, `announcers`, `queue`, `track`, `download`, `config`, `logger`, `api`, `version`.

## Interface-Driven Architecture

- Define interfaces in `types.go` within the domain package (see `internal/stage/types.go` for `Stager`, `internal/validators/types.go` for `Validator`, `internal/announcers/kafka.go` for `Announcer`).
- Keep data structs that cross package boundaries in the same `types.go` file alongside the interface.
- Implementations of an interface reside in sub-packages: `internal/stage/s3compat/` and `internal/stage/filebased/` both implement `stage.Stager`; `internal/validators/kafka/` implements `validators.Validator`.

## Fake / Test Doubles

- Place fakes in a file named `fake.go` in the same package as the interface they implement (e.g., `internal/stage/fake.go`, `internal/validators/fake.go`, `internal/announcers/fake.go`).
- Name the struct `Fake` and include boolean tracking fields (e.g., `StageCalledV`, `StatusCalledV`) with thread-safe accessor methods using `sync.Mutex`.
- Use `ShouldError bool` on fakes to toggle error paths in tests.

## Handler Construction Pattern

- Construct HTTP handlers with a `NewHandler` function that accepts interface dependencies and returns `http.HandlerFunc`. See `upload.NewHandler`, `track.NewHandler`, `download.NewHandler`.
- Wire dependencies in `cmd/insights-ingress/main.go` using concrete implementations selected by config (e.g., `getStager` selects between `s3compat` and `filebased`).

## Import Conventions

- Alias `internal/logger` as `l` (e.g., `l "github.com/redhatinsights/insights-ingress-go/internal/logger"`).
- Alias `prometheus/client_golang/prometheus` as `p` and `prometheus/promauto` as `pa` in metrics files.
- Use dot-imports (`. "..."`) for Ginkgo (`github.com/onsi/ginkgo`) and Gomega (`github.com/onsi/gomega`) in test files, and for the package under test.

## Metrics

- Define Prometheus metrics as package-level `var` blocks using `promauto` (`pa.NewCounterVec`, `pa.NewHistogramVec`, `pa.NewGauge`).
- Place metrics declarations in a separate `metrics.go` file within the package (see `internal/upload/metrics.go`, `internal/validators/kafka/metrics.go`).
- Prefix metric names with `ingress_` (e.g., `ingress_requests`, `ingress_kafka_produced`, `ingress_stage_seconds`).

## Routing

- Use `github.com/go-chi/chi/v5` for HTTP routing. Mount the API sub-router at both `/api/ingress/v1` and `/r/insights/platform/ingress/v1`.
- Run a separate metrics HTTP server on `MetricsPort` with a dedicated `chi.NewRouter`.
- Apply `request_id.ConfiguredRequestID("x-rh-insights-request-id")` middleware for request ID propagation.
- Use `identity.EnforceIdentityWithLogger` from `platform-go-middlewares/v2` when `cfg.Auth` is true.

## Embedded Assets

- Embed the OpenAPI spec using `//go:embed openapi.json` in `internal/api/api.go` as `var ApiSpec []byte`.

## Build

- Build command: `go build -o insights-ingress-go cmd/insights-ingress/main.go`.
- On macOS, use `-tags dynamic` for both build and test (required by confluent-kafka-go CGo bindings).

## Verification

```bash
# Confirm all packages compile
go build ./...

# Run all tests (use -tags dynamic on macOS)
go test -p 1 -v ./...

# Verify interface satisfaction
grep -rn 'type Fake struct' internal/

# Check that metrics follow naming convention
grep -rn 'Name:.*"ingress_' internal/

# Confirm no exported packages outside cmd/ and internal/
ls -d */ | grep -v -E '^(cmd|internal|development|deploy|docs|dashboards|licenses|\.)'
```
