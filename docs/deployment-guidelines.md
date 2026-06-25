# Deployment

## Container Images

- Use multi-stage builds: `registry.access.redhat.com/ubi9/go-toolset:latest` for building, `registry.access.redhat.com/ubi9/ubi-minimal:latest` for runtime.
- `Dockerfile` is the production image used by Tekton/Konflux pipelines. `Dockerfile.upstream` is for community/upstream builds via GitHub Actions. Keep them in sync except for the license copy path.
- Run as non-root user `1001` in the final stage.
- Build the binary with `go build -o insights-ingress-go cmd/insights-ingress/main.go` -- the entrypoint is `CMD ["/insights-ingress-go"]`.

## ClowdApp Manifest (deploy/clowdapp.yaml)

- The manifest is an OpenShift Template wrapping a `ClowdApp` of kind `cloud.redhat.com/v1alpha1`, named `ingress`.
- The single deployment is named `service` with a public web service on API path `ingress`.
- Health probes (liveness and readiness) use HTTP GET on `/` port `8000`, with `initialDelaySeconds: 35` and `timeoutSeconds: 120`.
- Mount an `emptyDir` volume named `tmpdir` at `/tmp` inside the pod.
- Declare two Kafka topics: `platform.payload-status` (3 partitions) and `platform.upload.announce` (64 partitions), each with 3 replicas.
- Declare `objectStore` referencing the `INGRESS_STAGEBUCKET` parameter (default `insights-upload-perma`).
- Hard dependencies: `puptoo`, `storage-broker`. Optional dependencies: `host-inventory`, `payload-tracker`.
- When adding new env vars, prefix with `INGRESS_` and add a matching `parameters` entry with a sensible default.
- `INGRESS_VALID_UPLOAD_TYPES` is a comma-separated allowlist -- append new types to the existing list.
- `INGRESS_MAXSIZEMAP` is a JSON map of per-type size overrides (e.g., `{"qpc": "157286400"}`).
- Resource defaults: CPU 200m/500m request/limit, memory 256Mi/512Mi request/limit.

## Tekton / Konflux Pipelines (.tekton/)

- Four PipelineRun manifests exist, split by branch and event:
  - `ingress-pull-request.yaml` / `ingress-push.yaml` target the `master` branch.
  - `ingress-sc-pull-request.yaml` / `ingress-sc-push.yaml` target the `security-compliance` branch.
- Master-branch pipelines pin to a versioned Konflux pipeline ref. When bumping the pipeline version, update both `ingress-pull-request.yaml` and `ingress-push.yaml` to the same tag.
- Master-branch builds enable hermetic mode (`hermetic: "true"`) with Go module prefetch. SC builds omit these parameters.
- PR images include `image-expires-after: 5d`; push images do not expire.

## GitHub Actions Workflows (.github/workflows/)

- `pr.yml`: Runs `go test ./...` and validates `internal/api/openapi.json` with `openapi-spec-validator` on every PR.
- `container-publish.yaml`: Builds via `Dockerfile.upstream`, pushes to `quay.io/iop/ingress`, and signs with cosign on push to `master` or semver tags.
- `security-workflow-template.yml`: Invokes the reusable ConsoleDot platform security scan.
- `renovate-validator-mintmaker.yaml`: Validates `renovate.json` changes.

## CI Scripts (Bonfire / App-SRE)

- `pr_check.sh` bootstraps Bonfire CICD, runs unit tests (`unit_test.sh`), deploys to ephemeral environment, and executes IQE smoke tests with plugin `ingress`.
- `build_deploy.sh` builds and pushes to `quay.io/cloudservices/insights-ingress` using short git SHA as tag.
- `unit_test.sh` runs tests with `ACG_CONFIG` pointing to `cdappconfig.json` at repo root.

## Local Development

- `development/local-dev-start.yml` runs Kafka (KRaft mode), Zookeeper, and MinIO via podman-compose.
- Use `make start-api-dependencies` then `make run-api` for S3-backed local dev.
- Use `make start-filebased-api-dependencies` then `make run-filebased-api` for file-based storage.

## Verification

```bash
# Validate ClowdApp manifest is valid YAML
python3 -c "import yaml; yaml.safe_load(open('deploy/clowdapp.yaml'))"

# Check Dockerfile builds
docker build -f Dockerfile -t ingress-test .

# Verify Tekton PipelineRun manifests parse
for f in .tekton/*.yaml; do python3 -c "import yaml; yaml.safe_load(open('$f'))"; done

# Confirm unit tests pass with mock Clowder config
ACG_CONFIG="$(pwd)/cdappconfig.json" go test -v ./...

# Ensure Tekton pipelines reference consistent versions
grep 'pipelinesascode.tekton.dev/pipeline' .tekton/ingress-pull-request.yaml .tekton/ingress-push.yaml
```
