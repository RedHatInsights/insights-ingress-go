# FROM registry.access.redhat.com/ubi9/go-toolset:latest as builder
FROM --platform=linux/amd64 quay.io/hummingbird/go:latest AS builder

WORKDIR /go/src/app

COPY cmd cmd

COPY internal internal

COPY go.mod go.mod

COPY go.sum go.sum

COPY licenses licenses

USER 0

RUN go get -d ./... && \
  go build -o insights-ingress-go cmd/insights-ingress/main.go

RUN cp /go/src/app/insights-ingress-go /usr/bin/

# FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
FROM --platform=linux/amd64 quay.io/hummingbird/core-runtime:latest

USER 0

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

RUN mkdir -p /licenses
COPY --from=builder /go/src/app/licenses/LICENSE /licenses

USER 1001

CMD ["/insights-ingress-go"]
