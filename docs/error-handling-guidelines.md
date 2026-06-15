# Error Handling

## Logging

- Use the global `l.Log` logger from `internal/logger` (aliased as `l`), not `fmt.Println` or the standard `log` package.
- Log errors using `logrus.Fields` with `{"error": err}` as the field key for attaching error values.
- In HTTP handlers, define a local `logerr` closure that calls `requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)`, following the pattern in `internal/upload/upload.go` and `internal/track/track.go`.
- Attach `request_id`, `source_host`, `account`, and `org_id` to the request-scoped logger. Add fields progressively as they become available.
- Include `status_code` in log fields when logging HTTP error responses.
- Use `Error` level for operational failures (staging errors, marshal failures, kafka producer errors). Use `Info` level for expected rejections (deny-listed org, invalid content type, authorization failures). Use `Debug` level for non-critical parse failures like missing optional metadata.
- Reserve `Fatal` level for startup-blocking failures in `cmd/insights-ingress/main.go` only. Do not use `Fatal` in request handlers.

## HTTP Status Codes

The upload handler in `internal/upload/upload.go` uses these specific status codes:

- `200 OK` -- test/connectivity requests only (`isTestRequest`)
- `201 Created` -- advisor service uploads with no metadata
- `202 Accepted` -- all other successful uploads
- `400 Bad Request` -- missing `file`/`upload` multipart form field; invalid request ID in track handler
- `403 Forbidden` -- deny-listed org ID; unauthorized track requests
- `404 Not Found` -- no payload-tracker data for a request ID
- `413 Request Entity Too Large` -- payload exceeds `DefaultMaxSize` or service-specific `MaxSizeMap` limit
- `415 Unsupported Media Type` -- content type fails regex match or service not in valid upload types
- `500 Internal Server Error` -- staging failure, JSON marshal failure, payload-tracker communication failure

Preserve the `201` vs `202` distinction: return `201` only when `vr.Service == "advisor"` and metadata is nil.

## Error Construction

- Use `errors.New()` for static error messages.
- Use `fmt.Errorf()` with `%v` for wrapping context in most places. Use `%w` wrapping only in `cmd/insights-ingress/main.go` for stager initialization errors.
- In `internal/stage/s3compat/s3compat.go`, construct error messages by concatenating via `.Error()` into a new `errors.New()`.

## Error Propagation

- HTTP handler functions do not return errors. They write the status code and body directly, then `return`.
- Interface methods (`Stager.Stage`, `Stager.GetURL`, `Validator.ValidateService`) return `(value, error)` tuples. The caller checks and responds with the appropriate HTTP status.
- `Validator.Validate` does not return an error. If JSON marshaling fails inside `internal/validators/kafka/kafka.go`, the error is logged and the function returns silently.
- `Kafka.Status` in `internal/announcers/kafka.go` logs marshal errors and returns without propagating.
- Failed Kafka publishes are retried by re-enqueuing the message onto the channel in `internal/queue/queue.go`.
- Metadata parsing errors (`GetMetadata`) are logged at `Debug` level and do not abort the request.

## Error Response Bodies

- Write plain-text error messages to the response body using `w.Write([]byte(...))` for client-facing errors (400, 403, 413). Do not return JSON-formatted error bodies.
- For 415 and 500 errors, write only the status code header with no response body.
- For successful responses (200, 201, 202), set `Content-Type: application/json` and write the JSON `responseBody` struct.

## Panic Usage

- `panic` is used only in `internal/config/config.go` for Kafka CA certificate write failure. Do not use `panic` elsewhere.
- The `chi` `middleware.Recoverer` is registered in `cmd/insights-ingress/main.go` to catch panics in request handlers.

## Prometheus Error Metrics

- Increment `ingress_kafka_produce_failures` (in `internal/queue/queue.go`) when a Kafka publish fails.
- Track response codes via `ingress_responses` counter with `useragent` and `code` labels in `internal/upload/metrics.go`.

## Verification

```bash
# Confirm all error logs use logrus.Fields{"error": err} pattern
grep -rn '\.Error(' internal/ cmd/ --include="*.go" | grep -v _test.go

# Confirm no fmt.Println or log.Fatal usage outside main
grep -rn 'fmt\.Println\|log\.Fatal\|log\.Print' internal/ --include="*.go"

# Confirm panic is only in config.go
grep -rn 'panic(' internal/ cmd/ --include="*.go" | grep -v _test.go

# Run tests
go test -p 1 -v ./...
```
