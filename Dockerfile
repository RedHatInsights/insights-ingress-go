FROM registry.access.redhat.com/ubi8/go-toolset:latest as builder

WORKDIR /go/src/app

COPY cmd cmd

COPY internal internal

COPY go.mod go.mod

COPY go.sum go.sum

COPY licenses licenses

USER 0

RUN go get -d ./... && \
    go build -buildvcs=false -o insights-ingress-go ./cmd/insights-ingress

RUN cp /go/src/app/insights-ingress-go /usr/bin/

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

WORKDIR /

RUN microdnf update -y

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

COPY --from=builder /go/src/app/licenses/LICENSE .

USER 1001

CMD ["/insights-ingress-go"]
