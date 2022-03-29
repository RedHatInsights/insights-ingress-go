# Insights Ingress

Ingress is designed to receive payloads from clients and distribute them via a
Kafka message queue to other platform services.

## Details

Ingress is a component of cloud.redhat.com that allows for clients to upload data
to Red Hat. The service sites behind a 3Scale gateway that handles authentication,
routing, and assignment of unique ID to the upload.

Ingress has an interface into cloud storage to retain customer data. It also connects
to a Kafka message queue in order to notify services of new and available uploads
for processing.

The service runs inside Openshift Dedicated.

## How It Works

![UML](http://www.plantuml.com/plantuml/png/ZL8zRzj03DtrAmXrQO4ubIKv2JGO6JiL7Jpr0mPzefo3xqCzKWQS8F-zTvRaE4EBQX8V7n_9ntjamI23DQ3TFX1priTOAzsZ4r16avDtKCKA3Rs3vif8rNA2tg1qFjZReJSUsrkcSDIA75hAMXJS8HDmrLEmw9Bys6Mn7gMRgCTw_oIy61FGuoa9PMD-iPxw_Pqu4QwOwedK0Jfj25W_qmrCGq6SAaQMMeqWfvuoD3ApKPiXK0RnkoZECtxPRBu12yh0e7nByB5ULf_hvUfJHePfak11gLZsln9bKSPozxRfkDT4ZTMzTqoNzNvys9c1Vgsll6nWD7ss0iH7gzyC-STj6h2yJ_mZ6jsYn9hPfUoh5uAGl0RVmSNLbnoLyeEJl86yIDyol_dfSeL2UnzE2UwyFsEM1DFr8_Roce10lmTYsUesqNPbLH-wdUEZQGzjToxfWtRfYPb4y66Vg0cVfehe_BjD2umv_PmIPL4_f708vappbhPSRLEOuDrT7SN6zthkZanNq9Obn2NFLD6MMD3sYHSFL2oAQb6iDikxPdNVbAlRX-LTNTxVrwll-Mls6AytMFC7 "Ingress Processing Flow")

The Ingress workflow is as follows:

  - The source client sends a payload of a specific content type to cloud.redhat.com
  - Ingress discovers a validating service from the content type, uploads the file to
  cloud storage, and puts a message on a kafka topic for that service.

### Kafka Topics

Ingress produces to topics to alert services of a new upload. The first topic an
upload is advertised to is the one gathered from the content type.

    - Produce to topic derived from content type: `platform.upload.service-name`

### Content Type

Uploads coming into Ingress should have the following content type:

`application/vnd.redhat.<service-name>.filename+tgz`

The filename and file type may vary. The portion to note is the service name as 
this is where Ingress discovers the proper validating service and what topic to 
place the message on. 

Example:

  `application/vnd.redhat.advisor.example+tgz` => `platform.upload.advisor`

### Message Formats

All messages placed on the Kafka topic will contain JSON with the details for the 
upload. They will contain the following structure:

Validation Messages:

       {
           "account": <account number>,
           "category": <currently translates to filename>,
           "content_type": <full content type string from the client>,
           "request_id": <uuid for the payload>,
           "principal": <currently the org ID>,
           "service": <service the upload goes to>,
           "size": <filesize in bytes>,
           "url": <URL to download the file>,
           "id": <host based inventory id if available>,
           "b64_identity": <the base64 encoded identity of the sender>,
           "timestamp": <the time the upload was received>,
           "metadata": <will contain additional json related to the uploading host>
       }

Any apps that will perform the validation should send **all** of the data they
received in addition to a `validation` key that contains `success` or `failure`
depending on whether the payload passed validation. This data should be sent to 
the `platform.upload.validation` topic.

Expected Validation Message:
    
    {
        ...all data received by validating app
        "validation": <"success"/"failure">
    }

## Errors

Ingress will report HTTP errors back to the client if something goes wrong with the
initial upload. It will be the responsibility of the client to communicate that
connection problem back to the user via a log message or some other means.

The connection from the client to Ingress is closed as soon as the upload finishes.
Errors regarding anything beyond that point (cloud storage uploads, message queue errors)
will only be reported in Platform logs. If the expected data is not available in
cloud.redhat.com, the customer should engage with support.

## Development

#### Prerequisites

Golang >= 1.12

#### Launching the Service

Compile the source code into a go binary:

    $> go build

Launch the application

    $> ./insights-ingress-go

The server should now be available on TCP port 3000.

    $> curl http://localhost:3000/api/ingress/v1/version

#### The Docker Option

You can also build ingress using Docker/Podman with the provided Dockerfile.

    $> docker build . -t ingress:latest

### Podman Compose

See [instructions](https://github.com/RedHatInsights/insights-ingress-go/blob/master/development/README.md)

#### Uploading a File

Ingress expects to be behind a 3Scale gateway that provides some mandatory headers.
You can provide these headers manually with a curl command

        $> curl -F "file=@somefile.tar.gz;type=application/vnd.redhat.<service-name>.somefile+tgz" -H "x-rh-identity: <base64 string>" -H "x-rh-request_id: testtesttest" \
        http://localhost:3000/api/ingress/v1/upload

Note, that your service name needs to be in the `INGRESS_VALIDTOPICS` variable inside of the `.env` file.

For testing, the following base64 identity can be used:

    eyJpZGVudGl0eSI6IHsidHlwZSI6ICJVc2VyIiwgImFjY291bnRfbnVtYmVyIjogIjAwMDAwMDEiLCAib3JnX2lkIjogIjAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=

This decodes to:

    {"identity": {"type": "User", "account_number": "0000001", "org_id": "000001", "internal": {"org_id": "000001"}}}

#### Testing

Use `go test` to test the application

    $> go test ./...
