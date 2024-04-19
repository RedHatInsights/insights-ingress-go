include development/.env

BINARY=insights-ingress-go

.PHONY: $(BINARY)

build: $(BINARY)

$(BINARY):
	go build -o $(BINARY) cmd/insights-ingress/main.go

test:
	go test -p 1 -v ./...

run-api: $(BINARY)
	INGRESS_STAGEBUCKET=insights-upload-perma \
	INGRESS_VALID_UPLOAD_TYPES=advisor,qpc \
	OPENSHIFT_BUILD_COMMIT=somestring \
	INGRESS_MINIODEV=true \
	INGRESS_MINIOACCESSKEY=$(MINIO_ACCESS_KEY) \
	INGRESS_MINIOSECRETKEY=$(MINIO_SECRET_KEY) \
	INGRESS_KAFKA_BROKERS=localhost:29092 \
	INGRESS_MINIOENDPOINT=localhost:9000 \
	./$(BINARY)

run-upload-test:
	curl -v -F "file=@go.mod;type=application/vnd.redhat.advisor.tgz" \
	-H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMiIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9LCAidHlwZSI6ICJiYXNpYyJ9fQ==" \
	-H "x-rh-insights-request-id: 1234" \
	http://localhost:3000/api/ingress/v1/upload

start-api-dependencies:
	cd development/ && sh .env && podman-compose -f local-dev-start.yml up zookeeper kafka minio minio-createbucket

stop-api-dependencies:
	podman-compose -f development/local-dev-start.yml down
