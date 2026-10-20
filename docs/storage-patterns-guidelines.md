# Storage Patterns

## Stager Interface

- Implement new storage backends by satisfying the `stage.Stager` interface defined in `internal/stage/types.go`. Methods: `Stage(*Input) (string, error)` and `GetURL(requestID string) (string, error)`.
- `Stage` persists the payload and returns a URL; `GetURL` returns a retrieval URL for a previously staged payload.
- Populate `stage.Input` with `Payload` (an `io.ReadCloser`), `Key` (the request ID), `Account`, `OrgId`, and `Size`.
- Register new stager implementations in `getStager` in `cmd/insights-ingress/main.go`, keyed by `StagerImplementation` config.

## S3-Compatible Storage (s3compat)

- S3 backend lives in `internal/stage/s3compat/s3compat.go` and uses `github.com/minio/minio-go/v6`. Do not upgrade to v7 without updating all call sites.
- `S3Stager` holds a `Bucket` string and a `*minio.Client`. Initialize via `s3compat.GetClient`, which reads from `config.StorageCfg`.
- When `StorageEndpoint` is empty, client defaults to `s3.amazonaws.com`. When `StorageRegion` is set, use `minio.NewWithRegion`; otherwise use `minio.New`.
- `Stage` uploads with content type hardcoded to `"application/gzip"` and attaches `requestID`, `account`, and `org` as S3 user metadata.
- `GetURL` generates presigned GET URLs with a 24-hour TTL (`time.Second * 24 * 60 * 60`).
- Default staging bucket name is `"available"`. In Clowder, bucket name comes from `clowder.ObjectBuckets`.

## File-Based Storage (filebased)

- File-based backend lives in `internal/stage/filebased/filebased.go`. Selected when `StagerImplementation` equals `"filebased"`.
- `FileBasedStager` requires `StagePath` (directory) and `BaseURL` (service URL prefix).
- Stored files named `<sanitized-request-id>.tar.gz`. The `filterAlphanumeric` function strips all non-letter/non-digit characters from the request ID.
- `GetURL` constructs download URLs as `<BaseURL>/download/<requestID>`.
- Download handler registered at `/download/{requestID}` only when stager is `"filebased"`.
- Download endpoint in `internal/download/download.go` sets `Content-Disposition` and `Content-Type: application/gzip`, serves via `http.ServeFile`.

## Configuration Environment Variables

- Storage config uses `INGRESS_` prefix. Key variables: `INGRESS_STAGERIMPLEMENTATION`, `INGRESS_STAGEBUCKET`, `INGRESS_MINIOENDPOINT`, `INGRESS_MINIOACCESSKEY`, `INGRESS_MINIOSECRETKEY`, `INGRESS_USESSL`, `INGRESS_STORAGEREGION`, `INGRESS_STORAGEFILESYSTEMPATH`, `INGRESS_SERVICEBASEURL`.
- Config maps `MinioEndpoint`/`MinioAccessKey`/`MinioSecretKey` viper keys to `StorageEndpoint`/`StorageAccessKey`/`StorageSecretKey` struct fields. Use `Minio*` key names in env vars.

## Testing Fakes

- Use `stage.Fake` from `internal/stage/fake.go` to mock the stager. Set `ShouldError` for failure paths. Access `StageCalled()` and `GetURLCalled()` (thread-safe via mutex).
- File-based tests use `t.TempDir()` or `os.MkdirTemp` for storage directory.

## Upload Flow and Stage Integration

- Upload handler constructs `stage.Input` using request ID as `Key`, multipart file as `Payload`, and identity-derived `Account`/`OrgId`/`Size`.
- After `stager.Stage` returns, handler calls `stageInput.Close()` explicitly. Redundant deferred close is intentional.
- Stage duration tracked via `ingress_stage_seconds` Prometheus histogram.
- Staged URL set on `validators.Request.URL` for downstream consumption.

## Local Development

- S3 mode: `make run-api` with MinIO via `make start-api-dependencies`. Compose file creates `insights-upload-perma` bucket.
- File-based mode: `make run-filebased-api` sets `INGRESS_STAGERIMPLEMENTATION=filebased`.

## Verification

```bash
# Confirm Stager interface is satisfied by both implementations
grep -n "stage.Stager" internal/stage/s3compat/s3compat.go cmd/insights-ingress/main.go

# Confirm presigned URL TTL is 24 hours
grep -n "time.Second\*24" internal/stage/s3compat/s3compat.go

# Confirm filebased filename sanitization
grep -n "filterAlphanumeric\|\.tar\.gz" internal/stage/filebased/filebased.go

# Confirm both stager implementations are wired in getStager
grep -n "StagerImplementation" cmd/insights-ingress/main.go

# Run tests
go test -p 1 -v ./internal/stage/... ./internal/upload/... ./internal/download/...
```
