################################
# STEP 1 build executable binary
################################
FROM registry.access.redhat.com/hi/go:latest-fips-builder AS builder

USER 0

WORKDIR /workspace

# Cache deps before copying source so that we do not need to re-download for every build
COPY go.mod go.sum .

# Fetch dependencies
RUN go mod download

# Now copy the rest of the files for build
COPY cmd cmd
COPY internal internal

# Build the binary
RUN go build -ldflags "-w -s" -o insights-ingress-go cmd/insights-ingress/main.go

############################
# STEP 2 build a small image
############################
FROM registry.access.redhat.com/hi/go:latest-fips

WORKDIR /

COPY --from=builder /workspace/insights-ingress-go /usr/bin/insights-ingress-go

COPY licenses/LICENSE /licenses/LICENSE

USER 1001

CMD ["insights-ingress-go"]
