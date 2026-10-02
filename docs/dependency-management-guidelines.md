# Dependency Management

## Module Configuration

- Use the module path `github.com/redhatinsights/insights-ingress-go` as declared in `go.mod`.
- Do not add `replace` directives to `go.mod`; the project relies on upstream module paths exclusively.
- Do not vendor dependencies; no `vendor/` directory exists. The project uses Go module proxy resolution.
- Commit both `go.mod` and `go.sum` together when changing dependencies.

## Automated Dependency Updates

- Renovate is configured in `renovate.jsonc` and extends the shared preset at `github>RedHatInsights/konflux-pipelines//renovate/foreman_satellite/renovate.json`. Do not add inline package rules that duplicate the shared preset.
- The shared preset owns the `foreman-*`/`SATELLITE-*` branches. Every inline `packageRule` in `renovate.jsonc` must set `"matchBaseBranches": ["master"]` so it cannot interfere with the preset's handling of those branches.
- Dependabot is no longer used; it was scoped solely to Docker base image updates on the `security-compliance` branch, which Renovate now covers. Dependabot alerts remain enabled, and Renovate's vulnerability fix PRs are built from them.
- A Renovate config validator workflow (`.github/workflows/renovate-validator-mintmaker.yaml`) runs on changes to `renovate.jsonc`.

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
- Base image updates, `security-compliance` branch included, are automated by Renovate.

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
