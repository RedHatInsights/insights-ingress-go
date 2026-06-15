# Dependency Management

## Module Configuration

- Use the module path `github.com/redhatinsights/insights-ingress-go` as declared in `go.mod`.
- Do not add `replace` directives to `go.mod`; the project relies on upstream module paths exclusively.
- Do not vendor dependencies; no `vendor/` directory exists. The project uses Go module proxy resolution.
- Commit both `go.mod` and `go.sum` together when changing dependencies.

## Automated Dependency Updates

- Renovate is configured in `renovate.json` and extends the shared preset at `github>RedHatInsights/konflux-pipelines//renovate/foreman_satellite/renovate.json`. Do not add inline package rules that duplicate the shared preset.
- Dependabot (`.github/dependabot.yml`) is scoped solely to Docker base image updates on the `security-compliance` branch. Do not add `gomod` ecosystem entries to Dependabot; Go dependency updates are handled by Renovate.
- A Renovate config validator workflow (`.github/workflows/renovate-validator-mintmaker.yaml`) runs on changes to `renovate.json`.

## Key Direct Dependencies

- **Kafka**: `github.com/confluentinc/confluent-kafka-go/v2` (cgo-based via librdkafka). On macOS, build and test with `-tags dynamic`. Production Docker builds on Linux do not require this tag.
- **S3/Object Storage**: `github.com/minio/minio-go/v6` for S3-compatible storage. Do not introduce `aws-sdk-go` for S3 operations; `aws-sdk-go` is used only in `internal/logger/logger.go` for CloudWatch credentials.
- **HTTP Router**: `github.com/go-chi/chi/v5`. Do not introduce alternative routers.
- **Configuration**: `github.com/spf13/viper`.
- **Logging**: `github.com/sirupsen/logrus` as the sole logging library.
- **Red Hat Platform Libraries**: `github.com/redhatinsights/app-common-go` (Clowder) and `github.com/redhatinsights/platform-go-middlewares/v2` (identity/request-id middleware).
- **Testing**: `github.com/onsi/ginkgo` (v1) and `github.com/onsi/gomega` for BDD-style tests. `github.com/jarcoal/httpmock` for HTTP mocking.
- **Metrics**: `github.com/prometheus/client_golang` for Prometheus instrumentation.

## Adding or Upgrading Dependencies

- Prefer upgrading existing dependencies over introducing alternatives that serve the same purpose.
- When adding a new direct dependency, ensure it appears under the `require` block (not just `indirect`). Run `go mod tidy` to clean up.
- Avoid dependencies that require CGO beyond `confluent-kafka-go`, since the Docker build uses `ubi9/go-toolset` (builder) and `ubi9/ubi-minimal` (runtime with no C toolchain).

## Hermetic / Konflux Builds

- Tekton pipelines in `.tekton/` build with `hermetic: "true"` and prefetch Go modules via `prefetch-input: '[{"type": "gomod", "path": "."}]'`. Adding non-Go dependency types requires updating the `prefetch-input` parameter in all four Tekton PipelineRun YAMLs.
- The Dockerfile does not run `go mod download` as a separate layer; dependencies are fetched implicitly during `go build`.

## Docker Base Images

- Builder stage: `registry.access.redhat.com/ubi9/go-toolset:latest`
- Runtime stage: `registry.access.redhat.com/ubi9/ubi-minimal:latest`
- Base image updates on `security-compliance` branch are automated by Dependabot.

## Licenses

- Place license files in the `licenses/` directory. The Dockerfiles copy `licenses/LICENSE` into the final image.

## Verification

```bash
# Ensure go.mod and go.sum are tidy
go mod tidy && git diff --exit-code go.mod go.sum

# Build on macOS (requires -tags dynamic)
make build

# Run tests
make test

# Check for unintended replace directives
grep '^replace' go.mod  # should produce no output

# Check for vendor directory (should not exist)
test ! -d vendor
```
