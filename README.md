# Insights Ingress

Ingress is designed to recieve payloads from clients and distribute them via a 
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

![UML](http://www.plantuml.com/plantuml/png/ZLD1Rziy3BtxLn3vBj-0DckQj8TW28gTsijQhBbrHS38TA9DbZI9pf0D_kyJPSTrCNIOWI3Iu-FZeoJUHCR0JMr0srsW60kVzbffZvP16KsMNq7pgD3G61eo4rNp4Rn1hboefuqt3ijff73GYYpMhzFsMrsKoBZ5I13ddaADLifrLSzNNQbbqezwj-TutWN0ur64Yov-lkhhlqti2IEcsfFw1fKs157_f3FeJK9ocNOrbHg1ZvuAD7nYepPDe0BITr8SFDwkrmyG6Rc9e5n9yFzYDd-_c5szAyX4wYLYerHA-rU9oulBb6vVEktwwgafspiRQMZlwR-jQUXvDPobKBjBE1q5i8Cupqtf2cfYb1j8NfHfIYf7naJEDy6R99XkQWaFzuzh4FOIddvDAbGS9qiOhQAhQPDtRTi-PwcKE98PJlzpxnogu6gu_NYNoPyS4nYg65mbcIyyASEEqQGoixClTa8Xk215BsGdfYRPLJwz0T-xo6dzGVutNwEpy4Fp7hB5i-6nR7IPDkb7hAQhzhbzmymZaLW5z7eQFIceN83Q1OAI6BMHzpzwQd-PWYNKSIStSK2Za_cK0tsuo7M369F2lPhq7-XxGv6JszJI1BUgd5tE5nFf4vLoZMN1Bz8tow0FsigW6O65UdMTyUtr8cbqeoeXcRuHj8aSKjLCxJq9wq-dcQ6GQUT25ih3D00IRK8k7kSRMGaYioRGO9rrJP6nzeBUFTx3EW4vqlMfxG5qAMyL3wWDRaqNFRrSUNTjMzoGBPlQ_0O0 "Ingress Processing Flow")

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

    $> curl http://localhost:3000/version

#### Uploading a File

Ingress expects to be behind a 3Scale gateway that provides some manadatory headers.
You can provide these headers manually with a curl command

    $> curl -F "file=@somefile.tar.gz" -H "x-rh-identity-header: <base64 string"> -H "x-rh-request_id: testtesttest" \
    http://localhost:3000/api/ingress/v1/upload

For testing, the following base64 identity can be used:

    eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=

This decodes to:

    {"identity": {"account_number": "0000001", "internal": {"org_id": "000001"}}}

#### Testing

Use `go test` to test the application

    $> go test ./...
