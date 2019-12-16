FROM golang:1.12

ENV GO111MODULE="on"

WORKDIR /go/src/app
COPY . .

RUN go get -d ./...
RUN go install -v ./...

CMD ["insights-ingress-go"]
