FROM registry.redhat.io/ubi8/go-toolset as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go get -d ./... && \
    go build -o insights-ingress-go cmd/insights-ingress/main.go

RUN cp /go/src/app/insights-ingress-go /usr/bin/ && \
    cp /go/src/app/api/openapi.json /var/tmp/

FROM registry.redhat.io/ubi8/ubi-minimal

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go
COPY --from=builder /go/src/app/api/openapi.json /var/tmp

USER 1001

CMD ["/insights-ingress-go"]
