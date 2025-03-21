include development/.env

BINARY=insights-ingress-go
DATA_DIR=$(PWD)/development/filebased
TAGS_OPTS=
BUILD_OPTS=
ifeq ($(shell uname -s), Darwin)
	TAGS_OPTS=-tags dynamic
	BUILD_OPTS=-ldflags -s -tags dynamic
endif

.PHONY: $(BINARY)

build: $(BINARY)

$(BINARY):
	go build ${BUILD_OPTS} -o $(BINARY) cmd/insights-ingress/main.go

test:
	go test ${TAGS_OPTS} -p 1 -v ./...

setup-filebased:
	mkdir -p $(DATA_DIR)

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

run-filebased-api: $(BINARY) setup-filebased
	INGRESS_STAGERIMPLEMENTATION="filebased" \
	INGRESS_STORAGEFILESYSTEMPATH=$(DATA_DIR) \
	INGRESS_VALID_UPLOAD_TYPES=advisor,qpc \
	OPENSHIFT_BUILD_COMMIT=somestring \
	INGRESS_KAFKA_BROKERS=localhost:29092 \
	./$(BINARY)

run-upload-test:
	curl -v -F "file=@go.mod;type=application/vnd.redhat.advisor.tgz" \
	-H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMiIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9LCAidHlwZSI6ICJiYXNpYyJ9fQ==" \
	-H "x-rh-insights-request-id: 1234" \
	http://localhost:3000/api/ingress/v1/upload

start-api-dependencies:
	cd development/ && sh .env && podman-compose -f $(PWD)/development/local-dev-start.yml up zookeeper kafka minio minio-createbucket

stop-api-dependencies:
	podman-compose -f $(PWD)/development/local-dev-start.yml down


start-filebased-api-dependencies:
	cd development/ && sh .env && podman-compose -f $(PWD)/development/local-dev-start.yml up zookeeper kafka
