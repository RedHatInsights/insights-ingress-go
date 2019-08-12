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

![UML](http://www.plantuml.com/plantuml/png/ZLLDRnmt3BtFho2cbpRmXzVhvB08ZA2z99V6gDrJOYou6kwk4esaIfnjjyNyzr8QpNgCPjmCO5iYdvuUAP9-5na3Twq1RNU1OgoyxBNI7Ys3CfeiFpCjeq93KzFff40r7y4RvAqBxKNdZSFc8b8uQ4KMMvg37D3e1bax-upOT_xPVh_HSmnuG6rm8yg41pSO2UBIKsZHoecfCT0NKanDDGHtVZj4j98mejxjEPuF3l1uJDJLu3-_BM7E0mjWWbHxKbzXgmr1r7_J6PHSW2H3TYqr6e6FdYeqFA8ba2vG1VAT64UDxnyUxY0oSXT1kORWnvl5yl9cyVgdYaoaGX4xfUJOzr9SNrtBSViKwH1NWSffxsoaKtYVVjYOZXvl9_bTmV0COog0dMGw1wMt4YPZUW3HeapNK3CAAtI-2zu8eJpl2ku-tZz0Gg_Wqp_r5XN7UWMJLNGjhVEkx_l7J2K79pIdxFzMbCCsk1RU__mXWtzrJ61eo-2swUGJBbt3Zj78DOkppxQc45n8brwbNH9LPrKvUoVxtaMMzTlqT_qbEWj_Qjv3bdXsAfQrRcGZFyJguhP_x6V4t1CbgO1UNqHF2gJAeNM1e256RLARJKhjXMPRGHjtwIN6xh8xAEugtnj4MBwuiANu0_tHKMHHAo7Lk57DudfvKwSuKIdNMKxsA_aMsUY3jgaxJJ8dwEitsRvvSoCSGxCcLsg-YMaTESYj6Leq9TH_bMQ4GgQT2vaeDzDeoDukvClZSYshR4czDkh91jOjHSRM9-k7-u_m88Ri6UBKzW2wIRLwcYFPErPkghk-hrv8jhn4vwLwVxy3-aikCbI9eLXBM1I1zpJsIF7FJi9LwRo6OYweiQilou0eFLpDCqm6KPYsmcn1Z8KuJ_a_9V84JAuYkBwiY-IwJoFXs-DfTg2l09i31TQHKZ6Fk7nwfenVHMm9MbdZZW0ZFEdwR5-LAHG1xL6u6vtiQFBGcSrFcgxV6irlnt5u_cmS_kBySJeCntywiEdKL-8fmsIUZgZSuk_aLzUQVm40 "Ingress Processing Flow")

The Ingress workflow is as follows:

  - The source client sends a payload of a specific content type to cloud.redhat.com
  - Ingress discovers a validating service from the content type, uploads the file to
  cloud storage, and puts a message on a kafka topic for that service.
  - The validating service, if there is one, checks the payload is safe and properly
  formatted.
  - The validating service then returns a message via the validation topic to 
  ingress with a failure or success message
  - If validation success, the upload is advertised to the rest of the platform.
  If it fails, the upload is put in rejected storage. This upload is retained for a
  period in the event that more diagnostics are necessary to discover why it failed.

### Kafka Topics

Ingress produces to topics to alert services of a new upload. The first topic an
upload is advertised to is the one gathered from the content type.

    - Produce to topic derived from content type: `platform.upload.service-name`
    - Consume from validation topic: `platform.upload.validation`
    - Produce to available topic: `platform.upload.available`

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

Available Messages:

      {
          "account": <account number>,
          "request_id": <uuid for the payload>,
          "principal": <currently the org ID>,
          "service": <the service that validated the payload>,
          "url": <URL to download the file>,
          "b64_identity": <the base64 encoded identity of the sender>,
          "id": <host based inventory id if available>, **RELOCATING TO EXTRAS**
          "satellite_managed": <boolean if the system is managed by satellite>, **RELOCATING TO EXTRAS**
          "timestamp": <the time the available message was put on the topic>,
          "extras": {
              "satellite_managed": <same as above>
              "id": <same as above>
              ...
          }
      }

Any apps that will perform the validation should send back all of the data they
received in addition to a `validation` key that contains `success` or `failure`
depending on whether the payload passed validation.

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
