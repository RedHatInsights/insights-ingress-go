/*
 * Requires: https://github.com/RedHatInsights/insights-pipeline-lib
 */

@Library("github.com/RedHatInsights/insights-pipeline-lib") _


if (env.CHANGE_ID) {
    runSmokeTest (
        ocDeployerBuilderPath: "platform/ingress",
        ocDeployerComponentPath: "platform/ingress",
        ocDeployerServiceSets: "advisor,platform,platform-mq",
        iqePlugins: ["iqe-advisor-plugin", "iqe-upload-plugin", "iqe-host-inventory-plugin"],
        pytestMarker: "smoke",
    )
}
