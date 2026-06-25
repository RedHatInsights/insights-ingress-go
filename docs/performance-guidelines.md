# Performance

## Upload Memory and Size Limits

- Set `MaxUploadMem` via `INGRESS_MAXUPLOADMEM` to control how much multipart form data is buffered in memory before spilling to disk. Default is 8 MB. Passed to `r.ParseMultipartForm` in `internal/upload/upload.go`.
- Enforce per-service upload size limits by populating `MaxSizeMap` (env var `INGRESS_MAXSIZEMAP`) rather than raising `DefaultMaxSize` (default 100 MB).

## Kafka Producer Concurrency

- `queue.Producer` in `internal/queue/queue.go` spawns an unbounded goroutine for every message read from the input channel. The `producerCount` gauge (`ingress_kafka_producer_go_routine_count`) tracks active producer goroutines.
- Validation producer channel is buffered at 100. Status announcer channel is buffered at 1000. Writes block when full, applying backpressure to the upload handler. Do not reduce these buffer sizes without understanding downstream throughput.
- Failed Kafka messages are re-enqueued to the same input channel with no backoff or attempt limit.

## HTTP Server Configuration

- The `http.Server` instances in `cmd/insights-ingress/main.go` do not set `ReadTimeout`, `WriteTimeout`, `ReadHeaderTimeout`, or `IdleTimeout`. When modifying server configuration, prefer setting these timeouts.
- `HTTPClientTimeout` (env var `INGRESS_HTTPCLIENTTIMEOUT`, default 10 seconds) controls the timeout for outbound HTTP client used by the track endpoint.

## Resource Cleanup

- Upload handler uses dual-close for the uploaded file: `defer file.Close()` as safety net, and `stageInput.Close()` after staging completes.
- Kafka producer (`p.Close()`) is deferred in `queue.Producer`. Use `Kafka.Stop()` (which calls `close(k.In)`) for graceful shutdown.

## Metrics Cardinality

- User-Agent strings are normalized via `NormalizeUserAgent` in `internal/upload/metrics.go` before use as Prometheus labels. When adding new user-agent-labeled metrics, use `NormalizeUserAgent` consistently.
- Avoid adding high-cardinality labels (request IDs, account numbers) to Prometheus metrics.
- Histogram buckets for `ingress_payload_sizes` are 10 KB, 100 KB, 1 MB, and 10 MB.

## Staging Performance

- `ingress_stage_seconds` histogram tracks S3/filesystem staging latency.
- Neither stager implementation streams data concurrently; staging is synchronous within the request handler.

## Profiling

- Set `INGRESS_PROFILE=true` to mount pprof endpoints at `/debug` via `middleware.Profiler()`. Keep disabled in production.

## Graceful Shutdown

- Main function uses `idleConnsClosed` channel and `signal.Notify` for SIGINT-based graceful shutdown. The shutdown has no timeout context, meaning it waits indefinitely for active connections to drain.

## Deny List Optimization

- Org ID deny list is converted from a slice to `map[string]bool` at handler construction time via `isRequestFromDenyListedOrgID`, not per-request. Preserve this closure-based approach.

## Test Constraints

- Tests run with `-p 1` (sequential, single process) to prevent race conditions from shared global state (Prometheus metric registrations via `promauto`).

## Verification

```bash
# Check for unbounded goroutine spawning patterns
grep -rn "go func" internal/ --include="*.go"

# Verify channel buffer sizes
grep -rn "make(chan" internal/ --include="*.go"

# Confirm no HTTP server timeouts are set
grep -rn "ReadTimeout\|WriteTimeout\|IdleTimeout\|ReadHeaderTimeout" cmd/ internal/ --include="*.go"

# Verify MaxUploadMem is passed to ParseMultipartForm
grep -rn "ParseMultipartForm" internal/ --include="*.go"

# Confirm profiler is gated behind config flag
grep -rn "Profile\|Profiler" cmd/ --include="*.go"

# Run tests
make test
```
