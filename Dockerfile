FROM golang:1.12

WORKDIR /go/src/app
COPY . .

RUN go get -d ./...
RUN go install -v ./...

CMD ["app"]
