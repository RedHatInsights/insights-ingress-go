# Local Development

The podman compose file found in this directory will standup kafka and minio for local use.

## Requirements
* podman
* podman-compose
* golang => 1.12

Docker will alos work if that is all you have

## Running

In order to run the local development stack, execute the following command. Substitute docker if needed

    $> podman-compose -f local-dev-start.yml up

Once the local environment is running, you must setup the buckets in Minio so that ingress has something to write to.

1. Login to localhost:9000 using the creds found in `.env`
2. Click the plus sign in the bottom right corner
3. Click create bucket
4. Create a bucket called `insights-upload-perma`
5. Once created, over over the name in the left navigation menu and click the 3 dots.
6. Click "edit policy"
7. Change the dropdown box to "Read and Write"

## Building and Running Ingress from Source

In the root of the insights-ingress-go repo, execute the following commands:

    $> go get ./...
    $> go build

There should now be an executable in the directory. You need to supply some env vars for the application to work
properly. Here is an example:

    $> INGRESS_STAGEBUCKET=insights-upload-perma INGRESS_VALIDTOPICS=advisor OPENSHIFT_BUILD_COMMIT=somestring INGRESS_MINIODEV=true INGRESS_MINIOACCESSKEY=$MINIO_ACCESS_KEY INGRESS_MINIOSECRETKEY=$MINIO_SECRET_KEY INGRESS_MINIOENDPOINT=localhost:9000 ./insights-ingress-go

## Running from Podman

In order to run as part of the dev environment, you can also uncomment the ingress stanza in `local-dev-start.yml`. This will attempt to pull the ingress image from Quay. You must first login to the Quay registry at quay.io and have access to the Cloudservices namespace. If you do not have this, please put in an RHIOPS ticket to gain access.

## Uploading

To test an upload, you can use an insights-archive, or any other type of file that is accepted by the end services of cloud.redhat.com

To generate an archive, install the insights-client on a rhel7 machine and run `insights-client --register --no-upload`. The client will create an archive and put it in a directory for you. You can keep this archive and upload with curl as described in the [main readme](https://github.com/RedHatInsights/insights-ingress-go#uploading-a-file)
