# Ingress Health and Readiness Requirements

## Goal

Implement differentiated liveness and readiness monitoring for Ingress that
reflects whether it can accept and stage uploads.

## Health endpoints

### `GET /healthz`

- Unauthenticated.
- Lightweight and local-only.
- Performs no external dependency checks.
- Returns HTTP `200` when the process is serving requests.

### `GET /status/`

- Unauthenticated.
- Checks Kafka broker metadata access.
- Checks the active staging backend:
  - S3/MinIO bucket access in S3 mode.
  - Filesystem staging directory access in filesystem mode.
- Runs checks with bounded timeouts.
- Returns HTTP `200` only when all checks pass.
- Returns HTTP `503` when any critical check fails.
- Returns structured JSON with per-dependency status.
- Does not expose credentials or sensitive connection details.

Example response:

```json
{
  "status": "ok",
  "dependencies": {
    "kafka": {"status": "ok"},
    "storage": {"status": "ok", "backend": "s3"}
  }
}
```

## Kubernetes probes

- Configure `livenessProbe` to use `/healthz`.
- Configure `readinessProbe` to use `/status/`.
- Use the deployment's web port.
- Use short, bounded probe timeouts.
- Suggested defaults:
  - `periodSeconds: 10`
  - `timeoutSeconds: 2–5`
  - `failureThreshold: 3`
- Add a `startupProbe` if startup timing requires additional protection.

## Acceptance criteria

- `/healthz` returns `200` without contacting dependencies.
- `/status/` returns `200` when Kafka and the active storage backend are available.
- Kafka failure causes `/status/` to return `503`.
- S3/MinIO failure causes `/status/` to return `503` in S3 mode.
- Filesystem failure causes `/status/` to return `503` in filesystem mode.
- Filesystem mode checks the filesystem backend rather than S3.
- The response includes per-dependency statuses without secrets.
- Dependency failures are logged with structured fields.
- Prometheus exposes dependency health metrics.
- Kubernetes manifests use separate liveness and readiness paths.
- Unit tests cover healthy and failed dependency checks.
- README or operations documentation explains the endpoints and failure behavior.
- Local verification covers Kafka, MinIO/S3, and filesystem storage modes.
