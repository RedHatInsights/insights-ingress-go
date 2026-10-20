# Data Validation

## Content-Type Validation

- Parse the file part's `Content-Type` header (from the multipart MIME header, not the request header) using `getServiceDescriptor()` in `internal/upload/validation.go`.
- Accept three legacy gzip content types as hardcoded aliases mapping to service `"advisor"` / category `"upload"`:
  - `application/x-gzip; charset=binary`
  - `application/gzip`
  - `application/gzip; charset=binary`
- For all other content types, match against the regex `application/vnd\.redhat\.([a-z0-9-]+)\.([a-z0-9-]+).*` where capture group 1 is the service name and group 2 is the category.
- Return HTTP 415 when the content type matches neither the legacy aliases nor the regex pattern.
- After extracting the `ServiceDescriptor`, call `validator.ValidateService()` to verify the service name exists in the `validUploadTypes` map. Return HTTP 415 if the service is not recognized.

## Multipart Form Handling

- Accept the payload file from either a `"file"` or `"upload"` form field name, tried in that order via `GetFile()` in `internal/upload/upload.go`. Return HTTP 400 if neither is found.
- Pass `cfg.MaxUploadMem` (default 8 MB) to `r.ParseMultipartForm()` to control memory threshold before writing to disk.
- Read metadata from a `"metadata"` form part, trying `r.FormFile("metadata")` first (as a file part), then `r.FormValue("metadata")` (as a form value). Metadata is optional; failure to read it is logged at Debug level and does not reject the request.

## Metadata Schema

- Deserialize metadata JSON into the `validators.Metadata` struct defined in `internal/validators/types.go`.
- After unmarshaling, set `md.Reporter = "ingress"` and `md.StaleTimestamp = time.Now().AddDate(0, 0, 30)` -- these are overwritten regardless of client input.
- All `Metadata` fields are optional (`omitempty` JSON tags) except `Reporter` and `StaleTimestamp`, which ingress populates.
- The `QueueKey` field in metadata, when present, is used as the Kafka message key for partition routing.
- Invalid metadata JSON causes metadata to be silently dropped, but the upload still proceeds.

## File Size Validation

- Check per-service size limits first via `cfg.MaxSizeMap[vr.Service]`, then fall back to `cfg.DefaultMaxSize` (default 100 MB).
- `MaxSizeMap` values are strings parsed to `int64` with `strconv.ParseInt`. Provide values in bytes as strings (e.g., `"157286400"` for 150 MB).
- Return HTTP 413 when the file exceeds the applicable limit.

## Response Status Codes

- Return 201 only when the service is `"advisor"` AND no metadata was provided.
- Return 202 for all other successful uploads.
- Return 200 for test/connection-check requests.
- Return 403 for deny-listed org IDs.

## Test Request Detection

- `isTestRequest()` in `internal/upload/upload.go` detects two test patterns:
  1. Form value `test=test` (current clients).
  2. JSON body `{"test": "test"}` with `Content-Type: application/json` (legacy/satellite clients), matched via regex.
- Test requests bypass all file and content-type validation and return HTTP 200.

## Org ID Fallback

- When `id.Identity.OrgID` is empty but `id.Identity.Internal.OrgID` is set, copy the internal value to the top-level `OrgID`. This fallback is applied before deny-list checking.

## Validator Interface

- Implement the `validators.Validator` interface (defined in `internal/validators/types.go`) with `Validate(*Request)` and `ValidateService(*ServiceDescriptor) error`.
- `ValidateService` is the synchronous gate that rejects unknown services before staging. `Validate` is the async step that publishes to Kafka after staging.
- Use `validators.Fake` for testing; it returns an error only when `service.Service == "failed"`.

## Valid Upload Types Configuration

- Configured via `INGRESS_VALID_UPLOAD_TYPES` as a comma-separated list (default: `"unit,announce"`).
- Parsed in `internal/config/config.go` with `strings.Split(options.GetString("Valid_Upload_Types"), ",")`.
- Passed as variadic args to `kafka.New()` which builds a `map[string]bool` lookup.

## Verification

```bash
# Run all tests
go test -p 1 -v ./...

# Run upload validation tests only
go test -v ./internal/upload/...

# Verify the content-type regex compiles
grep -n 'contentTypePat' internal/upload/validation.go

# Check valid upload types default
grep -n 'Valid_Upload_Types' internal/config/config.go

# List all Fake test doubles
grep -rn 'type Fake struct' internal/
```
