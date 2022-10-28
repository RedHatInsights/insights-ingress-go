build:

	go build -o insights-ingress-go cmd/insights-ingress/main.go

test:

	go test -p 1 -v ./...
