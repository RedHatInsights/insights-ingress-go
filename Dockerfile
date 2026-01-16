FROM registry.access.redhat.com/ubi9/go-toolset:latest@sha256:a532ce56e98300a4594b25f8df35016d55de69e4df00062b8e04b3840511e494 as builder

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

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7-1768783948

WORKDIR /

COPY --from=builder /go/src/app/insights-ingress-go ./insights-ingress-go

RUN mkdir -p /licenses
COPY --from=builder /go/src/app/licenses/LICENSE /licenses

USER 1001

CMD ["/insights-ingress-go"]

# Define labels for the ingress
LABEL url="https://www.redhat.com"
LABEL name="ingress" \
      description="This adds the satellite/ingress-rhel9 image to the Red Hat container registry. To pull this container image, run the following command: podman pull registry.stage.redhat.io/satellite/ingress-rhel9" \
      summary="A new satellite/ingress-rhel9 container image is now available as a Technology Preview in the Red Hat container registry."
LABEL com.redhat.component="ingress" \
      io.k8s.display-name="IoP ingress" \
      io.k8s.description="This adds the satellite/ingress image to the Red Hat container registry. To pull this container image, run the following command: podman pull registry.stage.redhat.io/satellite/ingress-rhel9" \
      io.openshift.tags="insights satellite iop ingress"
