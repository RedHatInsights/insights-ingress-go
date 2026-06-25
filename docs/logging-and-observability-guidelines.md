# Logging and Observability

## Logger Access

- Import the logger package as `l "github.com/redhatinsights/insights-ingress-go/internal/logger"` and use the global `l.Log` singleton. Do not create additional `logrus.Logger` instances.
- Initialize the logger once via `l.InitLogger(cfg)` in `cmd/insights-ingress/main.go` before any log calls.
- In test suites, call `l.InitLogger(config.Get())` in each Ginkgo `TestXxx` function before `RunSpecs`.

## Structured Logging with logrus

- Use `l.Log.WithFields(logrus.Fields{...})` for all log calls. Prefer structured fields over string interpolation.
- In HTTP handlers, create a `requestLogger` by enriching `l.Log` with baseline fields, then chain additional fields as context grows:
```go
requestLogger := l.Log.WithFields(logrus.Fields{
    "request_id":  reqID,
    "source_host": cfg.Hostname,
    "name":        "ingress",
})
```
- Define a local `logerr` closure for repeated error logging within a handler:
```go
logerr := func(msg string, err error) {
    requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
}
```
- Pass errors in fields using the key `"error"`, not as part of the message string.
- Include `"request_id"` in fields for any log entry related to a specific request.

## Log Levels

- `DEBUG` -- internal processing details (metadata parse failures, topic posts, stage durations).
- `INFO` -- request lifecycle events (payload received, sent to validation, denied uploads, startup).
- `ERROR` -- actionable failures (Kafka producer errors, marshal failures, identity decode failures).
- `Fatal` -- only for unrecoverable startup/shutdown failures.
- Log level set via `INGRESS_LOGLEVEL` env var (values: `DEBUG`, `ERROR`, default `INFO`). During tests, log level is forced to `FatalLevel`.

## CloudWatch Formatter

- The `CustomCloudwatch` formatter in `internal/logger/logger.go` outputs JSON with fixed fields: `@timestamp`, `@version`, `message`, `levelname`, `source_host`, `app` (hardcoded to `"ingress"`), and `caller`.
- `ReportCaller` is enabled on the logger.
- CloudWatch hook is attached only when `AwsAccessKeyId` is non-empty. Uses `lc.NewBatchWriterWithDuration` from `platform-go-middlewares/v2/logging/cloudwatch` with a 10-second batch interval.

## Request ID Tracing

- Request IDs propagated via `x-rh-insights-request-id` HTTP header, configured with `request_id.ConfiguredRequestID("x-rh-insights-request-id")` middleware.
- Retrieve with `request_id.GetReqID(r.Context())`.
- Used as storage object key, in Kafka messages, payload tracker status, and response bodies.

## Prometheus Metrics

- Use `prometheus/client_golang/prometheus/promauto` (aliased `pa`) for metric registration. Metrics auto-register with the default registry.
- Prefix all metric names with `ingress_`. Existing metrics:
  - `internal/upload/metrics.go`: `ingress_requests`, `ingress_payload_sizes`, `ingress_stage_seconds`, `ingress_responses`
  - `internal/validators/kafka/metrics.go`: `ingress_processed_payloads`, `ingress_validate_elapsed_seconds`, `ingress_message_produced`
  - `internal/queue/queue.go`: `ingress_kafka_produced`, `ingress_publish_seconds`, `ingress_kafka_produce_failures`, `ingress_kafka_producer_go_routine_count`
  - `internal/version/version.go`: `ingress_version` (created via `Namespace: "ingress"` + `Name: "version"`)
- Declare metrics as package-level `var` blocks in dedicated `metrics.go` files. Keep metric helper functions unexported and co-located.
- Normalize user-agent strings through `NormalizeUserAgent()` in `internal/upload/metrics.go` to prevent high-cardinality label explosion.
- Metrics endpoint served on separate HTTP server on `MetricsPort` (default 8080) via `promhttp.Handler()` at `/metrics`.

## Response Metrics Middleware

- Wrap upload routes with `upload.ResponseMetricsMiddleware` to track response codes per user-agent via `metricTrackingResponseWriter`.

## Grafana Dashboard

- Dashboard definition at `dashboards/grafana-dashboard-insights-ingress-general.configmap.yaml`. When adding new metrics, verify they are referenced in this dashboard.

## Verification

```bash
# Confirm all metrics use the ingress_ prefix
grep -rn 'Name:' internal/upload/metrics.go internal/validators/kafka/metrics.go internal/queue/queue.go internal/version/version.go | grep -v 'ingress_'

# Confirm logger is imported with the l alias
grep -rn 'l "github.com/redhatinsights/insights-ingress-go/internal/logger"' internal/ cmd/

# Confirm request_id field is present in handler log entries
grep -rn 'request_id' internal/upload/upload.go internal/track/track.go

# Confirm test suites initialize the logger
grep -rn 'l\.InitLogger' internal/ --include="*_suite_test.go"
```
