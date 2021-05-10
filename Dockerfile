FROM registry.redhat.io/ubi8/go-toolset as builder

USER 0
WORKDIR /go/src/app
COPY go.mod go.sum .
RUN go mod download
COPY main.go .
COPY pkg/ ./pkg/
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o insights-ingress-go main.go

FROM registry.redhat.io/ubi8/go-toolset
WORKDIR /
COPY --from=builder /go/src/app/insights-ingress-go /insights-ingress-go
COPY openapi.json /var/tmp
USER 1001

ENTRYPOINT ["/insights-ingress-go"]
