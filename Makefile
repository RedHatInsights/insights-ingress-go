build:

	go build -o insights-ingress cmd/insights-ingress/main.go

test:

	go test -p 1 -v ./...