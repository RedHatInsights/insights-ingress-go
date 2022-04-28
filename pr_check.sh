#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ingress"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="ingress"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/insights-ingress"  

IQE_PLUGINS="ingress"
IQE_MARKER_EXPRESSION="smoke"
IQE_FILTER_EXPRESSION=""
IQE_CJI_TIMEOUT="30m"

EXTRA_DEPLOY_ARGS="advisor"

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

source $CICD_ROOT/build.sh
source $APP_ROOT/unit_test.sh
source $CICD_ROOT/deploy_ephemeral_env.sh
source $CICD_ROOT/cji_smoke_test.sh
source $CICD_ROOT/post_test_results.sh
