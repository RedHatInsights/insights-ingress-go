FROM registry.access.redhat.com/ubi9/go-toolset:latest@sha256:2a6a828a4020b8a03875bb0fa72cf41c0bd8695c2e5c60f49acb5fcdbd1e35cf as builder

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

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7-1773204619

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

RUN mkdir -p /licenses
COPY --from=builder /go/src/app/licenses/LICENSE /licenses

USER 1001

CMD ["/insights-ingress-go"]
