FROM registry.access.redhat.com/ubi9/go-toolset:latest@sha256:ecfbfea4a8c479a9ae9655ef6bdb692fcdd2d1bdfcea86ad7b168de147a852a2 as builder

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

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.9-1795198179

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

RUN mkdir -p /licenses
COPY --from=builder /go/src/app/licenses/LICENSE /licenses

USER 1001

CMD ["/insights-ingress-go"]
