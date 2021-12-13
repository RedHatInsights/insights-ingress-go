FROM registry.redhat.io/ubi8/go-toolset:1.16.7 as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go get -d ./... && \
    go build -o insights-ingress-go cmd/insights-ingress/main.go

RUN cp /go/src/app/insights-ingress-go /usr/bin/

FROM registry.redhat.io/ubi8/ubi-minimal:8.5

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

USER 1001

CMD ["/insights-ingress-go"]
