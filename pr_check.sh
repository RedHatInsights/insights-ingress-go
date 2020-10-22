#!/bin/bash

make test 

if [ $? != 0 ]; then
    exit 1
fi

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ingress"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="ingress"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/insights-ingress-go"  # TODO: look IMAGE up from build_deploy.sh?


# ---------------------------
# We'll take it from here ...
# ---------------------------

set -ex

# TODO: custom jenkins agent image
export LANG=en_US.utf-8
export LC_ALL=en_US.utf-8

# TODO: check quay to see if image is already built -- also, merge this PR with master first?
BUILD_NEEDED=true

if [ $BUILD_NEEDED ]; then
    export PUSH_TO_LATEST=false
    source build_deploy.sh
fi

IMAGE_TAG=$(git rev-parse --short=7 HEAD)
GIT_COMMIT=$(git rev-parse HEAD)

# TODO: create custom jenkins agent image that has a lot of this stuff pre-installed
git clone https://github.com/RedHatInsights/bonfire.git
cd bonfire
python3 -m venv .venv
source .venv/bin/activate
pip install --upgrade pip setuptools wheel
pip install .

# TODO: get the value of this secret fixed ...
export QONTRACT_BASE_URL="https://$APP_INTERFACE_BASE_URL/graphql"
export QONTRACT_USERNAME=$APP_INTERFACE_USERNAME
export QONTRACT_PASSWORD=$APP_INTERFACE_PASSWORD

# Deploy configurations in OpenShift
oc login --token=$OC_LOGIN_TOKEN --server=$OC_LOGIN_SERVER

NAMESPACE=$(bonfire namespace reserve)

trap "bonfire namespace release $NAMESPACE" EXIT ERR SIGINT SIGTERM

# Get k8s resources for app and its dependencies (use insights-stage instead of insights-production for now)
# -> use this PR as the template ref when downloading configurations for this component
# -> use this PR's newly built image in the deployed configurations
bonfire config get \
    --ref-env insights-stage \
    --app $APP_NAME \
    --set-template-ref $COMPONENT_NAME=$GIT_COMMIT \
    --set-image-tag $IMAGE=$IMAGE_TAG \
    --get-dependencies \
    --namespace $NAMESPACE \
    > k8s_resources.json

oc apply -f k8s_resources.json -n $NAMESPACE

# Wait for everything to go 'active'
bonfire namespace wait-on-resources $NAMESPACE

# Spin up iqe pod
# python utils/create_iqe_pod.py $NAMESPACE
oc rsh -n $NAMESPACE iqe-tests curl -si http://ingress-ingress:8000/metrics
