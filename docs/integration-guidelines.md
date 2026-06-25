# Integration

## Storage (S3 / Filebased Staging)

- Implement new storage backends by satisfying the `stage.Stager` interface defined in `internal/stage/types.go` (methods: `Stage(*Input) (string, error)` and `GetURL(requestID string) (string, error)`).
- The S3-compatible client uses `minio/minio-go/v6` (not the AWS SDK) -- see `internal/stage/s3compat/s3compat.go`. Retain this library for S3 operations.
- When staging to S3, attach `requestID`, `account`, and `org` as `UserMetadata` on `minio.PutObjectOptions`. Content type is hardcoded to `application/gzip`.
- Presigned URLs use a 24-hour expiry via `PresignedGetObject`.
- Stager selection is controlled by `config.IngressConfig.StagerImplementation` -- only `"s3"` and `"filebased"` are valid. The `getStager` function in `cmd/insights-ingress/main.go` is the single decision point.

## Kafka

- Use `confluent-kafka-go/v2` for all Kafka production -- see `internal/queue/queue.go`. Do not introduce a second Kafka client library.
- The `queue.Producer` function consumes from a `chan validators.ValidationMessage` channel and produces to a single topic.
- On delivery failure, the message is re-enqueued to the same input channel.
- Two distinct Kafka producers exist at runtime: one for `platform.upload.announce`, one for `platform.payload-status`.
- Topic names are resolved through `config.GetTopic()` which uses Clowder's `KafkaTopics` mapping when enabled.

## Payload Tracker

- Status announcements go through the `announcers.Announcer` interface defined in `internal/announcers/kafka.go`.
- Two status events are emitted per successful upload: `"received"` (before staging) and `"success"` (after staging).
- The `/track/{requestID}` endpoint in `internal/track/track.go` proxies to the payload-tracker HTTP service at `PayloadTrackerURL`.
- Track endpoint validates request ID as UUID via `google/uuid`.
- Authorization checks `OrgID` match, but bypasses for `Associate` identity type and trusted integration test certificates.

## Platform Middlewares

- Use `platform-go-middlewares/v2` (not v1). Import path: `github.com/redhatinsights/platform-go-middlewares/v2/identity` and `.../request_id`.
- Request IDs extracted via `request_id.ConfiguredRequestID("x-rh-insights-request-id")` -- this custom header name is required.
- Identity retrieved with `identity.GetIdentity(r.Context())` returning `identity.XRHID`.
- When `OrgID` is empty but `Internal.OrgID` is populated, copy internal to top-level.
- Auth enforcement uses `identity.EnforceIdentityWithLogger(identityErrorLogFunc)`.

## Content Types and Validation

- Content-type routing handled by `getServiceDescriptor` in `internal/upload/validation.go`.
- Legacy gzip types map to service `"advisor"`, category `"upload"`.
- Custom content types follow `application/vnd.redhat.{service}.{category}` pattern.
- Service name checked against `ValidUploadTypes` configured via `INGRESS_VALID_UPLOAD_TYPES`.

## Testing

- Tests use Ginkgo/Gomega. Each test package has a `*_suite_test.go` bootstrap file.
- Fake implementations: `stage.Fake`, `validators.Fake`, `announcers.Fake`. Use these instead of mocking.
- HTTP tests use `httptest.NewRecorder` with `identity.WithIdentity` for context injection.
- Track tests use `jarcoal/httpmock` to stub payload-tracker HTTP calls.

## Verification

```bash
# Run all tests
go test -p 1 -v ./...

# Build the binary
go build -o insights-ingress-go cmd/insights-ingress/main.go

# Verify all interfaces are satisfied
go vet ./...
```
