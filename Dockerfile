FROM registry.access.redhat.com/ubi8/go-toolset:1.20.12-5.1713833129 as builder

WORKDIR /go/src/app
COPY . .

USER 0

RUN go get -d ./... && \
    go build -o insights-ingress-go cmd/insights-ingress/main.go

RUN cp /go/src/app/insights-ingress-go /usr/bin/

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.9-1161

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

COPY --from=builder /go/src/app/licenses/LICENSE .

USER 1001

CMD ["/insights-ingress-go"]
