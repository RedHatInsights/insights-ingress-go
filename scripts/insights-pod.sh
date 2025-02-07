#!/usr/bin/env bash

set -e

PODNAME="insights"
WORKDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

INSIGHTS_PROXY_PORT=1337
COMPLIANCE_FRONTEND_PORT=8002

if ! podman pod exists $PODNAME; then
	podman pod create --name "$PODNAME" \
	-p "${INSIGHTS_PROXY_PORT}" \
	-p "${COMPLIANCE_FRONTEND_PORT}"
else
	echo "pod $PODNAME already exists, delete it first!" && exit 1
fi
