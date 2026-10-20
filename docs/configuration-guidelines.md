# Configuration

## Configuration Architecture

- Runtime configuration is centralized in `internal/config/config.go` via the `Get()` function, which returns a pointer to `IngressConfig`.
- Configuration uses two Viper instances: `options` (with env prefix `INGRESS`) for application config and `kubenv` (no prefix) for Kubernetes/OpenShift env vars like `OPENSHIFT_BUILD_COMMIT`.
- When Clowder is enabled (`clowder.IsClowderEnabled()`), Clowder's `LoadedConfig` overrides Viper defaults for Kafka, storage, ports, TLS, and CloudWatch settings.
- `config.Get()` creates a new Viper instance and re-reads all config on every call. Callers in `main.go` call it once and pass the resulting `IngressConfig` to downstream components.

## Adding New Configuration Values

- Register every new config key with `options.SetDefault()` in `Get()` before the Clowder block.
- Add a corresponding field to `IngressConfig` or one of its nested structs (`KafkaCfg`, `StorageCfg`, `LoggingCfg`, `KafkaSSLCfg`).
- Read the value from `options` using typed getters (`GetString`, `GetBool`, `GetInt`, `GetInt64`, `GetStringSlice`, `GetStringMapString`).
- If the value should be overridden by Clowder, add a second `options.SetDefault()` or `options.Set()` call inside the `if clowder.IsClowderEnabled()` block.

## Environment Variable Naming

- Viper env prefix is `INGRESS` (set via `options.SetEnvPrefix("INGRESS")`), so env vars follow `INGRESS_<KEY>`.
- Viper key names use mixed case (e.g., `KafkaBrokers`, `MinioEndpoint`, `Valid_Upload_Types`). Viper maps these case-insensitively to env vars.
- Keys containing underscores (e.g., `Valid_Upload_Types`, `Deny_Listed_OrgIDs`) are used as-is; do not normalize to camelCase.

## Clowder Integration

- Gate all Clowder-specific logic behind `clowder.IsClowderEnabled()`.
- Use `clowder.LoadedConfig` to access Clowder-provided values.
- Look up Kafka topic names via `clowder.KafkaTopics["<requested-topic-name>"].Name`.
- Use the `GetTopic()` helper in `internal/config/config.go` for runtime topic resolution.
- Access the storage bucket via `clowder.ObjectBuckets[sb]` where `sb` comes from `INGRESS_STAGEBUCKET`.
- Use `options.Set()` (not `SetDefault`) for Clowder values that should unconditionally override env vars.
- Use `options.SetDefault()` for Clowder values where env var overrides should still be possible.

## Clowder Config File

- `cdappconfig.json` at the repo root is a local Clowder config fixture for development/testing.
- The ClowdApp deployment manifest is at `deploy/clowdapp.yaml` -- declares Kafka topics, object store buckets, and all `INGRESS_*` env vars.

## Stager Implementation Selection

- `StagerImplementation` controls storage backend: `"s3"` (default) or `"filebased"`.
- Register new stagers by adding a corresponding branch in `getStager()` in `cmd/insights-ingress/main.go`. `GetStagerImplementation()` in `internal/config/config.go` normalizes the stager name but is not where new stagers are registered.

## Per-Service Max Size Overrides

- `DefaultMaxSize` sets the global upload size limit (default: 100 MB).
- `MaxSizeMap` is a JSON string-map providing per-service overrides (e.g., `{"qpc": "157286400"}`). Values are string representations of byte counts.
- Upload size checking in `internal/upload/upload.go` looks up `cfg.MaxSizeMap[vr.Service]` first, then falls back to `cfg.DefaultMaxSize`.

## Valid Upload Types

- `Valid_Upload_Types` is a comma-separated string split into a slice at startup.
- Controls which content-type service names are accepted. Unrecognized service names receive HTTP 415.
- Update `INGRESS_VALID_UPLOAD_TYPES` in `deploy/clowdapp.yaml` when deploying.

## Deny-Listed Org IDs

- `Deny_Listed_OrgIDs` is a string slice config key (env var `INGRESS_DENY_LISTED_ORGIDS`).
- Denied org IDs result in HTTP 403. The deny list is converted to a map at handler creation time.

## Debug Configuration

- `Debug` (bool) and `DebugUserAgent` (regex pattern) work together: when both are set, requests matching the user-agent regex get full request dumps logged.
- `DebugUserAgent` is compiled to a `*regexp.Regexp` at config load time.

## Local Development Configuration

- Use `development/.env` to set local env vars (sourced by the Makefile).
- `make run-api` for S3/MinIO-backed local development.
- `make run-filebased-api` for filesystem-backed local development.

## Verification

```bash
# Confirm all config keys have SetDefault calls
grep -c 'SetDefault' internal/config/config.go

# Confirm env prefix is set
grep 'SetEnvPrefix' internal/config/config.go

# Confirm Clowder guard pattern exists
grep 'IsClowderEnabled' internal/config/config.go

# Check clowdapp.yaml declares expected INGRESS_ env vars
grep 'INGRESS_' deploy/clowdapp.yaml

# Verify config is imported by downstream packages
grep -r 'internal/config' --include="*.go" cmd/ internal/
```
