FROM registry.redhat.io/ubi8/go-toolset

WORKDIR /go/src/app
COPY . .

RUN go get -d ./... && \
    go install -v ./...

USER 0
RUN cp /opt/app-root/src/go/bin/insights-ingress-go /usr/bin/ && \
    cp /go/src/app/openapi.json /var/tmp/

RUN yum remove -y kernel-headers npm nodejs nodejs-full-i18n

USER 1001
CMD ["insights-ingress-go"]
