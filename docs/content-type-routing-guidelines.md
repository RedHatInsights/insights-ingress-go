# Content-Type Routing

## Routing Architecture

Uploads route to a single Kafka topic (`platform.upload.announce`). Downstream consumers distinguish payloads by the `service` Kafka header attached to each message, not by separate topics per service.

## Content-Type Format and Parsing

- Use the vendor MIME type pattern: `application/vnd.redhat.<service>.<category>` where `<service>` and `<category>` each contain only lowercase alphanumeric characters and hyphens (`[a-z0-9-]+`).
- The regex in `internal/upload/validation.go` extracts `service` (capture group 1) and `category` (capture group 2) from the content type via `contentTypePat`:
  ```
  application/vnd\.redhat\.([a-z0-9-]+)\.([a-z0-9-]+).*
  ```
- Three legacy gzip content types are hardcoded to route to the `advisor` service with category `upload`:
  - `application/x-gzip; charset=binary`
  - `application/gzip`
  - `application/gzip; charset=binary`
- Any content type that does not match the vendor pattern or a legacy gzip variant results in HTTP 415.

## Service Validation via ValidUploadTypes

- `INGRESS_VALID_UPLOAD_TYPES` controls which service names are accepted. It is a comma-separated list.
- `kafka.Validator.ValidateService()` in `internal/validators/kafka/kafka.go` checks the service name against a `map[string]bool` built from that list. Unrecognized service returns HTTP 415.
- When adding a new upload type, add the service name to `INGRESS_VALID_UPLOAD_TYPES` in `deploy/clowdapp.yaml` and to `development/.env`.

## Kafka Message Production

- Validation messages are produced to the single announce topic (`platform.upload.announce` by default, remapped by Clowder when enabled).
- The `service` value is attached as a Kafka message header (`{"service": "<service>"}`) in `kafka.Validator.Validate()`.
- If upload metadata contains a `queue_key` field, it is used as the Kafka message key, enabling partition-based ordering.
- Status messages go to `platform.payload-status` via `announcers.Kafka.Status()`.

## Per-Service Max Size Overrides

- `INGRESS_MAXSIZEMAP` is a JSON object mapping service names to byte-size strings (e.g., `{"qpc": "157286400"}`). Checked in `upload.NewHandler()` before `DefaultMaxSize`.
- Exceeding the limit returns HTTP 413.

## Response Codes by Content Type

- `advisor` service uploads without metadata return HTTP 201.
- All other service uploads return HTTP 202.
- Invalid content type or unrecognized service: HTTP 415.
- File size exceeded: HTTP 413.

## ServiceDescriptor Struct

- `validators.ServiceDescriptor` in `internal/validators/types.go` carries `Service` and `Category` strings. It is the routing key used by `ValidateService()`.
- Populate both `Service` and `Category` from the content type; do not leave `Category` empty.

## Testing Content-Type Routing

- Use `validators.Fake` from `internal/validators/fake.go` in unit tests. It rejects service name `"failed"` and accepts everything else.
- Tests in `internal/upload/upload_test.go` use the `boiler` helper with `FilePart` structs that set `ContentType` on the multipart file header.
- Tests in `internal/validators/kafka/kafka_test.go` validate `ValidateService()` against a constructed valid-types list.

## Clowder Topic Remapping

- `config.GetTopic()` in `internal/config/config.go` translates a requested topic name to the Clowder-assigned name when `clowder.IsClowderEnabled()` is true. Use `config.GetTopic()` for any topic name resolution rather than hardcoding topic strings.

## Verification

```bash
# Confirm content type regex compiles and has two capture groups
grep -n 'contentTypePat' internal/upload/validation.go

# Confirm valid upload types are set in deployment config
grep 'INGRESS_VALID_UPLOAD_TYPES' deploy/clowdapp.yaml

# Confirm Kafka header is set with the service name
grep -A3 'Headers:' internal/validators/kafka/kafka.go

# Confirm MaxSizeMap usage for per-service overrides
grep -n 'MaxSizeMap' internal/upload/upload.go internal/config/config.go

# Run unit tests for upload and kafka validator
go test ./internal/upload/... ./internal/validators/kafka/...
```
