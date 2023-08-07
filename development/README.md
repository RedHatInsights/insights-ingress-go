# Local Development

The podman compose file found in this directory will standup kafka and minio for local use.

## Requirements
* podman
* podman-compose
* golang => 1.19

Docker will alos work if that is all you have

## Running

In order to run the local development stack, execute the following command. Substitute docker if needed

    $> podman-compose -f local-dev-start.yml up

This should build the ingress image, start the dependencies (kafka, minio, etc) and start the ingress
service within podman.

Once the local environment is running, you can load `http://localhost:9990/` to access the minio console.
The required buckets should be created automatically by podman.

## Building and Running Ingress from Source

It is also possible to run the ingress service locally but configure it so that the service uses the
dependencies that are running within podman.

The `make start-api-dependencies` Makefile rule can be used to start the dependencies using podman-compose:

    $> make start-api-dependencies

The `make run-api` Makefile rule can be used to run ingress locally:

    $> make run-api

## Uploading

The `make run-upload-test` Makefile rule can be used to run send an upload to the ingress service locally:

    $> make run-upload-test

To test an upload, you can use an insights-archive, or any other type of file that is accepted by the end services of cloud.redhat.com

To generate an archive, install the insights-client on a rhel7 machine and run `insights-client --register --no-upload`. The client will create an archive and put it in a directory for you. You can keep this archive and upload with curl as described in the [main readme](https://github.com/RedHatInsights/insights-ingress-go#uploading-a-file)
