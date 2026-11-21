################################
# STEP 1 build executable binary
################################
FROM registry.access.redhat.com/hi/go:latest-fips-builder@sha256:c52ae1e6fe4d3a0bbe3b4457eca4691f762c7f71605423a4069b1c3dbd7daccc AS builder

USER 0

WORKDIR /workspace

# Cache deps before copying source so that we do not need to re-download for every build
COPY go.mod go.sum .

# Fetch dependencies hermetically by sourcing the local prefetch cache
RUN if [ -f /cachi2/cachi2.env ]; then . /cachi2/cachi2.env; fi && \
    go mod download

# Now copy the rest of the files for build
COPY cmd cmd
COPY internal internal

# Build the binary using the local cached modules
RUN if [ -f /cachi2/cachi2.env ]; then . /cachi2/cachi2.env; fi && \
    go build -ldflags "-w -s" -o insights-ingress-go cmd/insights-ingress/main.go

############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/hi/go:latest-fips@sha256:ee1ec2fd8d76564b1d45ba348508db163b2d9d3bb9a9cec166b112ab87c2bbd8

WORKDIR /

COPY --from=builder /workspace/insights-ingress-go /usr/bin/insights-ingress-go

COPY licenses/LICENSE /licenses/LICENSE

ARG IMAGE_NAME
ARG VERSION
LABEL summary="Red Hat Insights Ingress" \
    description="Red Hat Insights Ingress service" \
    io.k8s.description="Red Hat Insights Ingress service" \
    io.k8s.display-name="Insights Ingress" \
    com.redhat.component="costmanagement-ingress-rhel9-container" \
    name="$IMAGE_NAME" \
    version="$VERSION" \
    release="1" \
    vendor="Red Hat, Inc." \
    distribution-scope="public" \
    cpe="cpe:/a:redhat:cost_management_on_premise:1::el9" \
    maintainer="Red Hat Cost Management Services <cost-mgmt@redhat.com>"

USER 1001

CMD ["insights-ingress-go"]
