#!/usr/bin/env bash

set -e

PODNAME="insights-ingress-go"
#PODNAME="insights"

WORKDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && cd .. && pwd )"

#CHECK_RESPONSE_SCRIPT='check-response-code.sh'
WAIT_FOR_CMD="${WORKDIR}/scripts/check-response-code.sh"

MINIO_PORT=9000
ZOOKEEPER_CLIENT_PORT=32181
KAFKA_ADVERTISED_LISTENERS_PORT=29092
INGRESS_PORT=3000
INGRESS_HOST_PORT=8080
#INSIGHTS_PROXY_URL="ci.foo.redhat.com:1337"
INSIGHTS_PROXY_URL="192.168.122.1:1337"

MINIO_DATA_DIR="${WORKDIR}/data"
MINIO_CONFIG_DIR="${WORKDIR}/config"
MINIO_ACCESS_KEY=BQA2GEXO711FVBVXDWKM
MINIO_SECRET_KEY=uvgz3LCwWM3e400cDkQIH/y1Y4xgU4iV91CwFSPC
INGRESS_VALID_TOPICS=testareno,advisor

#PODMAN_NETWORK="cni-podman1"
#PODMAN_GATEWAY=$(podman network inspect $PODMAN_NETWORK | jq -r '..| .gateway? // empty')

if podman pod exists $PODNAME; then
	podman pod rm -f $PODNAME
fi
#	echo "pod $PODNAME not found !" && exit 1
podman pod create --name "$PODNAME" \
    -p "${KAFKA_ADVERTISED_LISTENERS_PORT}:${KAFKA_ADVERTISED_LISTENERS_PORT}" \
    -p "${INGRESS_HOST_PORT}:${INGRESS_PORT}" \
    -p "${MINIO_PORT}:${MINIO_PORT}"
#	podman pod create --name "$PODNAME" --network "$PODMAN_NETWORK" -p "$MINIO_PORT" \
#		--add-host "ci.foo.redhat.com:$PODMAN_GATEWAY" \
#		--add-host "qa.foo.redhat.com:$PODMAN_GATEWAY" \
#		--add-host "stage.foo.redhat.com:$PODMAN_GATEWAY" \
#		--add-host "prod.foo.redhat.com:$PODMAN_GATEWAY"

#podman build . -t "ingress"

# zookeeper
podman run --pod "$PODNAME" -d --name "zookeeper" \
	-e ZOOKEEPER_CLIENT_PORT="${ZOOKEEPER_CLIENT_PORT}" \
	-e ZOOKEEPER_SERVER_ID=1 \
	confluentinc/cp-zookeeper

# kafka
podman run --pod "$PODNAME" -d --name "kafka" \
	-e KAFKA_ADVERTISED_LISTENERS="PLAINTEXT://192.168.122.1:${KAFKA_ADVERTISED_LISTENERS_PORT}" \
	-e KAFKA_BROKER_ID=1 \
	-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
	-e KAFKA_ZOOKEEPER_CONNECT="localhost:${ZOOKEEPER_CLIENT_PORT}" \
	confluentinc/cp-kafka

# minio
podman run --pod "$PODNAME" -d --name "minio" \
	-e MINIO_ACCESS_KEY="$MINIO_ACCESS_KEY" \
	-e MINIO_SECRET_KEY="$MINIO_SECRET_KEY" \
	-v "$MINIO_DATA_DIR:/data:z" \
	-v "$MINIO_CONFIG_DIR:/root/.minio:z" \
	minio/minio \
	server /data
#
#
#until podman run --rm --pod "$PODNAME" -v "${WORKDIR}/${CHECK_RESPONSE_SCRIPT}:/${CHECK_RESPONSE_SCRIPT}" "nginx:alpine" "/${CHECK_RESPONSE_SCRIPT}" "http://localhost:${MINIO_PORT}/minio/health/ready" 200 ; do
#until podman run --pod --rm -v "./scripts/check-response-code.sh:/check.sh:z" nginx:alpine /check.sh localhost:80 201
#	>&2 echo "Minio is not yet ready..."
#	sleep 1
#done
#
until $WAIT_FOR_CMD "http://localhost:${MINIO_PORT}/minio/health/ready" 200 ; do
	>&2 echo "Minio is not yet ready..."
	sleep 1
done

# createbuckets
podman run --pod "$PODNAME" -d --name "createbuckets" \
	-v "$MINIO_DATA_DIR:/data:z" \
	-v "$MINIO_CONFIG_DIR:/root/.minio:z" \
	-e MINIO_ACCESS_KEY="$MINIO_ACCESS_KEY" \
	-e MINIO_SECRET_KEY="$MINIO_SECRET_KEY" \
    --entrypoint "/bin/sh" \
	  minio/mc \
      -c \
	  "/usr/bin/mc config host add myminio http://localhost:${MINIO_PORT} $MINIO_ACCESS_KEY $MINIO_SECRET_KEY ;\
	  /usr/bin/mc mb myminio/insights-upload-perma;\
      /usr/bin/mc mb myminio/insights-upload-rejected;\
      /usr/bin/mc policy set download myminio/insights-upload-perma;\
      /usr/bin/mc policy set download myminio/insights-upload-rejected;\
      exit 0;"

# ingress
podman run --pod "$PODNAME" -d --name "ingress" \
	-v "$MINIO_DATA_DIR:/data:z" \
    -e AWS_ACCESS_KEY_ID=$MINIO_ACCESS_KEY \
	-e AWS_SECRET_ACCESS_KEY=$MINIO_SECRET_KEY \
    -e AWS_REGION=us-east-1 \
    -e INGRESS_STAGEBUCKET=insights-upload-perma \
    -e INGRESS_REJECTBUCKET=insights-upload-rejected \
    -e INGRESS_INVENTORYURL="https://${INSIGHTS_PROXY_URL}/api/inventory/v1/hosts" \
    -e INGRESS_VALIDTOPICS=$INGRESS_VALID_TOPICS \
    -e OPENSHIFT_BUILD_COMMIT=woopwoop \
    -e INGRESS_MINIODEV=true \
    -e INGRESS_MINIOACCESSKEY=$MINIO_ACCESS_KEY \
    -e INGRESS_MINIOSECRETKEY=$MINIO_SECRET_KEY \
    -e INGRESS_MINIOENDPOINT=localhost:${MINIO_PORT} \
	ingress:latest

