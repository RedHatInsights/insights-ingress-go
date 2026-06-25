# API Contracts

## Routing and URL Structure

- Mount all API routes under two base paths via `chi.Mount`: `/api/ingress/v1` and `/r/insights/platform/ingress/v1`. Both serve the same subrouter.
- Use `go-chi/chi/v5` for all routing. Do not introduce alternative routers.
- Define handlers as `http.HandlerFunc` values returned by constructor functions (e.g., `upload.NewHandler`, `track.NewHandler`, `download.NewHandler`).
- Extract path parameters with `chi.URLParam(r, "paramName")`, not from `mux.Vars` or manual parsing.

## Endpoints

The service exposes exactly these routes on the API subrouter:

| Method | Path                  | Handler                       | Auth Required |
|--------|-----------------------|-------------------------------|---------------|
| GET    | `/`                   | `lubDub` (health check)       | Conditional   |
| POST   | `/upload`             | `upload.NewHandler`           | Conditional   |
| GET    | `/track/{requestID}`  | `track.NewHandler`            | Required*     |
| GET    | `/version`            | `version.GetVersion`          | No            |
| GET    | `/openapi.json`       | `apiSpec` (serves embedded spec) | No         |

*The `/track/{requestID}` route is only registered when `cfg.Auth` is true. When auth is disabled, this endpoint does not exist.

A separate `/download/{requestID}` route is mounted at the root router level only when `StagerImplementation == "filebased"`.

Metrics are served on a separate port via `/metrics` (Prometheus `promhttp.Handler`).

## OpenAPI Specification

- The canonical API spec lives at `internal/api/openapi.json` as an OpenAPI 3.0.0 document.
- It is embedded into the binary at compile time via `//go:embed openapi.json` in `internal/api/api.go` and exposed as `api.ApiSpec`.
- Keep `openapi.json` in sync with any endpoint or response schema changes. The spec defines three paths: `/upload`, `/version`, and `/track/{request_id}`.

## Content-Type Handling

- Uploads use `multipart/form-data`. The file part is read from either the `file` or `upload` form field (tried in that order via `GetFile` in `internal/upload/upload.go`).
- The service descriptor (service name + category) is extracted from the inner Content-Type header of the file part, not the outer request Content-Type.
- Custom content types follow the pattern: `application/vnd.redhat.<service>.<category>...` parsed by the regex in `internal/upload/validation.go`:
  ```
  application/vnd\.redhat\.([a-z0-9-]+)\.([a-z0-9-]+).*
  ```
- Legacy gzip types (`application/x-gzip; charset=binary`, `application/gzip`, `application/gzip; charset=binary`) map to service `advisor` with category `upload`.
- Unrecognized content types return HTTP 415.

## Request Headers

- `x-rh-insights-request-id`: Extracted by `request_id.ConfiguredRequestID` middleware. Used as the payload's unique key throughout the pipeline.
- `x-rh-identity`: Base64-encoded identity JSON. Decoded by `platform-go-middlewares/v2/identity`. Read directly from the header for `B64Identity` field on the validation request.
- `User-Agent`: Logged and used for Prometheus metric labels. Normalize via `upload.NormalizeUserAgent` before using as a label value.

## Authentication Middleware

- When `cfg.Auth` is true, apply `identity.EnforceIdentityWithLogger` middleware to protected routes.
- Pass a custom error logging function of signature `func(ctx context.Context, rawId, msg string)` to `EnforceIdentityWithLogger`.
- The `/version` and `/openapi.json` endpoints do not require identity enforcement regardless of auth configuration.
- The `/track/{requestID}` endpoint is only registered when `cfg.Auth` is true.

## Response Conventions

- Responses from `/upload` use `application/json` Content-Type.
- Successful upload with `advisor` service and no metadata returns HTTP 201. Other successful uploads return HTTP 202.
- Test connection requests (form value `test=test` or JSON body `{"test":"test"}`) return HTTP 200.
- The upload JSON response body uses the `responseBody` struct shape: `{"request_id": "...", "upload": {"account_number": "...", "org_id": "..."}}`.
- The `/version` endpoint returns `{"version": "...", "commit": "..."}` with HTTP 200.
- The `/track/{requestID}` endpoint returns `MinimalStatus` by default; full `TrackerResponse` when query param `verbosity >= 2`.

## HTTP Status Codes

| Code | Meaning in this service |
|------|------------------------|
| 200  | Test connection / health check / track response |
| 201  | Advisor upload without metadata accepted |
| 202  | Payload accepted for processing |
| 400  | Missing file/upload part, or invalid request ID format |
| 403  | Org ID is deny-listed, or track authorization failure |
| 404  | Request ID not found in payload tracker |
| 413  | File exceeds `DefaultMaxSize` or per-service `MaxSizeMap` limit |
| 415  | Content type unrecognized or service not in valid upload types |
| 500  | Staging failure or internal marshaling error |

## Middleware Ordering

Apply middleware to the upload route in this order (set via `sub.With(...)`):
1. `upload.ResponseMetricsMiddleware` (wraps `ResponseWriter` for status code tracking)
2. `identity.EnforceIdentityWithLogger` (when auth is enabled)
3. `middleware.Logger` (chi request logging)

## Metadata Handling

- Metadata is optional. It is read from either a `metadata` form file part or a `metadata` form value.
- When present, it is unmarshaled into `validators.Metadata`. The `Reporter` field is hard-coded to `"ingress"` and `StaleTimestamp` is set to 30 days from now.
- Failed metadata parsing logs a debug message but does not fail the request.

## Verification

```bash
# Confirm the OpenAPI spec is valid JSON and has expected paths
python3 -c "import json; d=json.load(open('internal/api/openapi.json')); assert '/upload' in d['paths'] and '/version' in d['paths'] and '/track/{request_id}' in d['paths']"

# Confirm both mount paths exist
grep -n 'r.Mount' cmd/insights-ingress/main.go

# Confirm upload form field names
grep -n 'FormFile' internal/upload/upload.go

# Confirm response status codes in upload handler
grep -n 'WriteHeader' internal/upload/upload.go

# Run the test suite
go test -p 1 -v ./...
```
