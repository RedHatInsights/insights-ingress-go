FROM registry.redhat.io/ubi8/go-toolset as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go get -d ./... && \
    go install -v ./...

RUN cp /opt/app-root/src/go/bin/insights-ingress-go /usr/bin/ && \
    cp /go/src/app/openapi.json /var/tmp/

FROM registry.redhat.io/ubi8/ubi-minimal

WORKDIR /

COPY --from=builder /opt/app-root/src/go/bin/insights-ingress-go ./insights-ingress-go
COPY --from=builder /go/src/app/openapi.json /var/tmp

USER 1001

CMD ["/insights-ingress-go"]
