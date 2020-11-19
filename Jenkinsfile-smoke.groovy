/*
 * Requires: https://github.com/RedHatInsights/insights-pipeline-lib
 */

@Library("github.com/RedHatInsights/insights-pipeline-lib@v3") _

execSmokeTest (
    ocDeployerBuilderPath: "ingress/ingress",
    ocDeployerComponentPath: "ingress/ingress",
    ocDeployerServiceSets: "ingress,inventory,platform-mq",
    iqePlugins: ["iqe-ingress-plugin","iqe-e2e-plugin"],
    pytestMarker: "smoke"
)
