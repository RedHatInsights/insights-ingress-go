# Security

## Identity and Authentication

- Retrieve caller identity via `identity.GetIdentity(r.Context())` which returns `identity.XRHID`. Identity is placed into context by `identity.EnforceIdentityWithLogger` middleware from `platform-go-middlewares/v2/identity`.
- Guard all authenticated routes with `identity.EnforceIdentityWithLogger(identityErrorLogFunc)` as chi middleware. When `cfg.Auth` is `false`, identity middleware is skipped entirely.
- Check `cfg.Auth` before calling `identity.GetIdentity(r.Context())`. Both `internal/upload/upload.go` and `internal/track/track.go` gate identity extraction on `cfg.Auth == true`.
- When top-level `Identity.OrgID` is empty, fall back to `Identity.Internal.OrgID`. This fallback happens in `internal/upload/upload.go` and should be preserved on any new handler that consumes org identity.
- Forward raw base64 identity header (`x-rh-identity`) to downstream validators via `B64Identity` field on `validators.Request`.

## Org-ID Deny List

- Configured via `INGRESS_DENY_LISTED_ORGIDS`, stored as `[]string` in `config.IngressConfig.DenyListedOrgIDs`.
- Deny list check is implemented as a closure returned by `isRequestFromDenyListedOrgID` in `internal/upload/upload.go`. Converts slice to `map[string]bool` once at handler creation time.
- Deny-listed requests receive HTTP 403 with body `"Upload denied. Please contact Red Hat Support."`.
- Deny list check runs before file parsing and staging. Preserve this ordering to avoid unnecessary I/O for blocked orgs.

## Track Endpoint Authorization

- `/track/{requestID}` enforces authorization by comparing `id.Identity.OrgID` against the org ID from the payload-tracker response via `isIdAuthorized`.
- Three identity types bypass the org-ID ownership check:
  1. `Identity.Type == "Associate"` (internal Red Hat associate)
  2. Trusted integration test X.509 certificates, identified by matching `SubjectDN` against `AutomatedIntegrationTestCertSubjectStage` and `AutomatedIntegrationTestCertSubjectProd` constants
- Request IDs validated as UUIDs via `uuid.Parse` before any downstream call.

## Kafka TLS and SASL Configuration

- Kafka security protocol defaults to `"PLAINTEXT"`.
- When Clowder is enabled and `broker.Authtype` is set, SASL credentials are read from Clowder broker config and stored in `config.KafkaSSLCfg`.
- `queue.Producer` in `internal/queue/queue.go` conditionally sets `ssl.ca.location`, `security.protocol`, `sasl.mechanism`, `sasl.username`, and `sasl.password` only when corresponding config values are non-empty.
- Both the validator producer and status announcer producer receive SSL/SASL config. When modifying Kafka producer configuration, update both paths in `cmd/insights-ingress/main.go`.

## Object Storage TLS

- MinIO/S3-compatible client in `internal/stage/s3compat/s3compat.go` uses `StorageCfg.UseSSL` to toggle TLS. When Clowder is enabled, `UseSSL` is set from `cfg.ObjectStore.Tls`.

## Debug Mode

- When `cfg.Debug` is true and `User-Agent` matches `cfg.DebugUserAgent` (compiled regex), full HTTP request including body is dumped to logs via `httputil.DumpRequest(r, true)`.
- Avoid enabling `INGRESS_DEBUG` in production. Request dump includes `x-rh-identity` header and full payload body.

## Content-Type as Security Boundary

- Content-type validated against regex in `internal/upload/validation.go`. Extracted service name checked against `ValidUploadTypes`. Only explicitly allowed service types can submit payloads.

## File-Based Stager

- File-based stager sanitizes request IDs via `filterAlphanumeric`, stripping all non-letter/non-digit characters before constructing file paths.
- The `/download/{requestID}` endpoint is mounted without identity middleware. Downloads are unauthenticated when using file-based stager.

## Verification

```bash
# Confirm identity middleware is applied
grep -n "EnforceIdentityWithLogger" cmd/insights-ingress/main.go

# Confirm deny list check runs before file processing
grep -n "isRequestFromDenyListedOrgID\|isCustomerDenyListed" internal/upload/upload.go

# Confirm SASL/TLS config is propagated to both producer configs
grep -n "KafkaSSLConfig" cmd/insights-ingress/main.go

# Confirm track endpoint validates request ID as UUID
grep -n "isValidUUID\|uuid.Parse" internal/track/track.go

# Confirm file-based stager sanitizes request IDs
grep -n "filterAlphanumeric" internal/stage/filebased/filebased.go

# Run security-relevant tests
go test ./internal/upload/... ./internal/track/... ./internal/validators/kafka/...
```
