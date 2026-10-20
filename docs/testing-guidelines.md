# Testing

## Test Framework and Imports

- Use Ginkgo v1 (`github.com/onsi/ginkgo`) and Gomega (`github.com/onsi/gomega`) for all tests. Do not use Ginkgo v2.
- Dot-import both Ginkgo and Gomega in test files: `. "github.com/onsi/ginkgo"` and `. "github.com/onsi/gomega"`.
- Dot-import the package under test when tests are in an `_test` package.

## Suite Files

- Each tested package has a `*_suite_test.go` file containing a single `Test*` function that bootstraps the Ginkgo runner.
- Initialize configuration and logger in every suite bootstrap function before `RunSpecs`:
  ```go
  func TestUpload(t *testing.T) {
      cfg := config.Get()
      RegisterFailHandler(Fail)
      l.InitLogger(cfg)
      RunSpecs(t, "Upload Suite")
  }
  ```
- The logger initialization via `l.InitLogger(cfg)` is required to avoid nil pointer panics.

## Test Structure

- Use `Describe` at the top level to name the component or function under test.
- Use `Context` to describe the specific scenario or input condition.
- Use `It` to state the expected outcome.
- Prefer `BeforeEach` over shared setup at the `Describe` level for mutable state.

## Fakes and Mocking

- Use in-repo fake implementations for dependency injection:
  - `stage.Fake` (`internal/stage/fake.go`) for `Stager`. Set `ShouldError: true` to simulate failures.
  - `validators.Fake` (`internal/validators/fake.go`) for `Validator`. Inspect `validator.In` after request to assert on `validators.Request`.
  - `announcers.Fake` (`internal/announcers/fake.go`) for `Announcer`.
- For external HTTP dependencies (payload-tracker), use `github.com/jarcoal/httpmock`. Call `httpmock.Activate()` in `BeforeEach`.

## HTTP Handler Testing

- Use `net/http/httptest.NewRecorder()` to capture handler responses.
- Construct handlers via `NewHandler` factory functions, passing fakes and `config.IngressConfig`.
- Wrap upload handlers with `request_id.ConfiguredRequestID("x-rh-insights-request-id")` middleware when the handler reads request IDs from context.
- Set identity context using `identity.WithIdentity(ctx, identity.XRHID{...})` from `platform-go-middlewares/v2/identity`.
- For chi router URL params, inject via `chi.NewRouteContext()` and add to request context with `chi.RouteCtxKey`.

## Multipart Upload Requests

- Build multipart requests using a local helper function with `multipart.Writer` parts and explicit MIME headers.
- Use a `FilePart` struct with `Name`, `Content`, and `ContentType` fields.
- Accepted part names for file uploads are `"file"` and `"upload"`.

## Config Overrides in Tests

- Override config values by modifying `*config.IngressConfig` returned from `config.Get()` before passing to `NewHandler`. Do not set environment variables in tests.
- Example: set `cfg.DefaultMaxSize = 1` to test size limits, set `cfg.DenyListedOrgIDs` for deny-list tests.

## Assertions

- Use `Expect(rr.Code).To(Equal(http.StatusOK))` for status codes.
- Use `Expect(err).To(BeNil())` for error checks (not `Expect(err).ToNot(HaveOccurred())`; the codebase consistently uses `BeNil()`).
- Inspect fake state after handler execution: `Expect(stager.StageCalled()).To(BeTrue())`.

## Running Tests

- Run all tests sequentially (required): `go test -p 1 -v ./...`
- Run with race detection (CI mode): `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...`
- CI sets `ACG_CONFIG` pointing to `cdappconfig.json`: `ACG_CONFIG="$(pwd)/cdappconfig.json" go test ...`
- On macOS, build tags required: `go test -tags dynamic -p 1 -v ./...`

## Verification

```bash
# Run the full test suite
go test -p 1 -v ./...

# Confirm all suite files initialize config and logger
grep -rn "l.InitLogger" --include='*_suite_test.go' internal/

# Confirm Ginkgo v1 dot-imports in all test files
grep -rn '"github.com/onsi/ginkgo"' --include='*_test.go' internal/

# Verify fakes exist for all injected interfaces
ls internal/stage/fake.go internal/validators/fake.go internal/announcers/fake.go
```
