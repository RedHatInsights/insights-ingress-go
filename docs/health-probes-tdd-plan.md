# Ingress Health Probes TDD Plan

## Design

Introduce a small health package with:

1. A dependency checker interface and HTTP readiness handler.
2. A Kafka checker that fetches broker metadata without producing messages.
3. Storage checks implemented by the active S3 and filesystem stagers.
4. A liveness handler that performs no dependency work.
5. A Prometheus gauge for the last result of each dependency check.

The readiness handler runs checks concurrently under one request deadline,
returns generic failure details to callers, and writes detailed structured
errors to logs.

## Red: tests first

1. Add readiness handler tests for all dependencies healthy:
   - Assert HTTP `200`.
   - Assert JSON reports overall `ok`.
   - Assert Kafka and storage have `ok` statuses.
2. Add a test for a failed critical dependency:
   - Assert HTTP `503`.
   - Assert only the dependency status is exposed, not the underlying secret or
     connection detail.
3. Add a timeout test:
   - Use a checker that waits for context cancellation.
   - Assert the handler returns `503` within the configured deadline.
4. Add liveness endpoint tests:
   - Assert HTTP `200`.
   - Use checkers that would fail if invoked and assert they are not called.
5. Add storage-mode tests:
   - S3 checker verifies bucket access.
   - Filesystem checker verifies the staging directory.
6. Add deployment-manifest assertions for `/healthz` and `/status/`.

## Green: implementation

1. Implement the liveness handler with a minimal `200` response.
2. Implement the readiness response model and concurrent checker execution.
3. Add Kafka metadata checking using the existing Confluent Kafka client and
   security settings.
4. Add S3 bucket access checking using the existing MinIO client.
5. Add filesystem staging-directory checking.
6. Register the endpoints on the web router without authentication.
7. Wire the active stager and Kafka checker into the readiness handler.
8. Add dependency health metrics and structured failure logs.
9. Configure Kubernetes probes with bounded timing values.

## Refactor and verification

1. Keep dependency failure response messages generic and free of secrets.
2. Ensure the readiness timeout is shorter than the Kubernetes probe timeout.
3. Preserve the existing root health route for compatibility.
4. Run `gofmt` on changed Go files.
5. Run focused health and staging tests.
6. Run `go test -p 1 -v ./...`.
7. Validate `deploy/clowdapp.yaml` as YAML.
8. Verify locally with Kafka and MinIO, then with filesystem staging:
   - All active dependencies available → readiness `200`.
   - Kafka unavailable → readiness `503`.
   - Active storage unavailable → readiness `503`.
