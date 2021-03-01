import argparse
import os
import sys
import subprocess

from string import Template

TOKEN = os.environ.get("SONAR_TOKEN")
SONAR_CONFIG = ".sonar/sonar-scanner.properties"
DEFAULT_PROJECT_NAME = os.getcwd().split("/")[-1]

p = argparse.ArgumentParser(add_help=False)
p.add_argument("--sonar_path", "-s",
               default="~/bin/sonar/lib/sonar-scanner-cli-*.jar",
               help="path to sonar binary")
p.add_argument("--token", "-t", help="sonar authentication token", required=True)
p.add_argument("--host", "-h",
               default="https://sonarqube.corp.redhat.com",
               help="host to submit the scan. Defaults to Red Hat")
p.add_argument("--project", "-p",
               default=DEFAULT_PROJECT_NAME,
               help="The name of the project in sonarqube.")
args = p.parse_args()


def initConfig(config):
    configLines = []
    with open(config, "w") as f:
        with open(".sonar/sonar-scanner.properties.template", "r") as r:
            for line in r.readlines():
                t = Template(line)
                configLines.append(t.substitute(host=args.host, project=args.project, token=args.token))
        f.writelines(configLines)


def runSonar(sonar_path):
    try:
        print("Running scan...")
        output = subprocess.check_output(f"java -jar {sonar_path.strip()} -Dproject.settings={SONAR_CONFIG}",
                                shell=True,
                                stderr=subprocess.STDOUT)
        print(output.decode("utf-8"))
    except Exception as e:
        print("Sonar Scan failed:", e.output)


if __name__ == "__main__":
    try:
        q = subprocess.check_output('ls ' + args.sonar_path, shell=True)
        if ".jar" not in str(q):
            print("""Scanner jar is required. Download from here:
                https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.2.0.1873.zip
                Unzip then set with --sonar-path""")
            sys.exit(1)
        else:
            SONAR_PATH = q.decode("utf-8")
    except subprocess.CalledProcessError as e:
        print(e.output)
    if not os.path.isfile(SONAR_CONFIG):
        initConfig(SONAR_CONFIG)

    runSonar(SONAR_PATH)
