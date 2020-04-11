FROM golang:1.12 as builder

ENV GO111MODULE="on"

WORKDIR /go/src/app
COPY . .

RUN go get -d ./... && \
    go install -v ./...

FROM registry.access.redhat.com/ubi8-minimal

COPY --from=builder /go/bin/insights-ingress-go /usr/bin/

USER 1001

CMD ["insights-ingress-go"]
