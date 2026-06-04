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

![UML](https://www.plantuml.com/plantuml/png/ZPDBZzem4CVl-HIZS6cbXOOam8e3scDFLG-SoY8qSIR1me_KTf1On7UlJHwxAvMgEU7__EOnVviNwz2uLWhWgZPaRTJuCsUyGUM02KxAVP8oor3G9sd8z2Xt5sW4kaeREMiReR6SMJ9dpaYXf4S8AgLRnIWgqM61bi1c3Hc9AhJlffXkkjPhty_o-kZij0j0WvTG9UhYqqq_psEm1pwGx4Zi11KNlZD_eoVeXmQ5qfyabHp1NHgAKBY1HYvQGn7uRwmupFXzk_q9tblNMc2w9FYIpxDl-NpnDI9XgIzXMyPysl-MI9FKfwltJRkzUjHdDrfP6jVRJGhHqdupUXdGpl712d3QM_rko3_kRWtIre4_e-0bEfypkZG1GJMoYo_hZj4FxGXCS1vq1QF7rnWPqwroyHhY97pp-EbLnGmTrTfSWbnfVTSaEGnlmMlNMn0C_Mx9kWCl0rQaMNwg2cNFeZoLrJsbCLo51oa2e4qTqA3tCtfr-7a8wtGn_XO2QP8_XsDhn1tJaWusEuHZa8jbxejrJpV4mmFz81siywthE-guz5EYR0BdhokP9jaqMMpdo_LYjKuNi-VvCazNgtpnAxuzjdtuFuoU3yBW-2EFTqV2aepTm_KlYyDzyTkhsabFOqrxc5Wl0Lh0Gfzf4hsGAbif_W00 "Ingress Processing Flow")

The Ingress workflow is as follows:

  - The source client sends a payload of a specific content type to cloud.redhat.com
  - Ingress discovers a validating service from the content type, uploads the file to
  cloud storage, and puts a message on the announcement topic while applying a header
  to identify the destination service.

For the vast majority of upload types, the above process is accurate. For payloads from both
**ansible tower** and the **insights cluster operator**, the flow is a bit different. The key difference here is that it relies on the UHC Auth Proxy for authentication.

![UML](https://www.plantuml.com/plantuml/png/TP5FSvim4CNl-HHRNzBEDEP0Jpt5mTF6qwRrXFYSMSC6D0Y9QbTntKzV2Mpe4FV4_dc_vV6uPK4dljLNxvGfj2y9Qf6EFoU9myEoKbBxlMToXJL2HfQ5RPDEeudC3KkfrJx9FjriusZty3rfaOLS63rdWK1bo2sxUFygFuPD-pwy92e-mY8RgaKeDuPLLGl3puuSYdMB3sSzDjYY2ffLNqHrjlunxLCkK5EOfdaiudwrtS1N53hWSTBvkWYhtNq6AoyrR9tzVOpYs94HLQ0eQo0dzweAcZXbAaVCqUHGHMZNQOlbMp6BTLZHdRDD_udvqCCmY6IUdfeBSDhlUrCj_h46xhGj6ZWVcO17qbEEOq23gTxV_TFJDX-4puzZX6DKNwmxe2jXZqmbM0Daoiug8pDs33U6Duzg9Xbp6g-BFG-11-lpyoF3wHHgXyVuZ3WBLa43Uryq9F-dvwcJAQ4Dgp2CPzx-XM_uqk3fp8pkhJpOL_hNoBLrrMQTd38FbQDVdbWswsjG1cn7Xclr8XUStWOpljL_0G00)

### Announcement Topic

Ingress produces a message on the `platform.upload.announce` topic to signal services
that an upload has arrived. The announce message contains a header that specifies which
content-type the message contains. For example:

    - {"service": "advisor"}

These kafka headers allow the consumer to filter out the messages that do not belong
to them without having to extract the full JSON of the incoming message. No performance
impact has been observed relating to this method of filtering.

### Content Type

Uploads coming into Ingress should have the following content type:

`application/vnd.redhat.<service-name>.filename+tgz`

The filename and file type may vary. The portion to note is the service name as
this is where Ingress discovers the proper validating service and what header value
to apply to the `service` header.

Example:

  `application/vnd.redhat.advisor.example+tgz` => `{"service": "advisor"}`

### Message Formats

All messages placed on the Kafka topic will contain JSON with the details for the 
upload. They will contain the following structure:

Validation Messages:

       {
           "account": <account number>,
           "org_id": <org id>,
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

The `platform.upload.validation` topic is consumed and handled by [Storage Broker](https://www.github.com/redhatinsights/insights-storage-broker).

If the app needs to relay a failure back to the customer via the notification
service, they can do so by supplying additional data as noted below:

Expected Validation Message:
    
    {
        ...all data received by validating app
        "validation": <"success"/"failure">
        ## additional notification data below ##
        "reason": "some error message",
        "system_id": <if available>,
        "hostname": <if available>,
        "reporter": "name of reporting app",
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

Golang >= 1.21

**macOS additional dependencies (for running with `-tags dynamic`):**

```bash
brew install pkg-config librdkafka
```

These are required for building with the `dynamic` tag, which links against the system librdkafka library for Kafka support.

#### Launching the Service

Compile the source code into a go binary:

    $> make build

Launch the application

    $> ./insights-ingress-go

The server should now be available on TCP port 3000.

    $> curl http://localhost:3000/api/ingress/v1/version

#### The Docker Option

You can also build ingress using Docker/Podman with the provided Dockerfile.

    $> docker build . -t ingress:latest

### Local Development

More information on local development can be found [here](./development/README.md)

#### Uploading a File

Ingress expects to be behind a 3Scale gateway that provides some mandatory headers.
You can provide these headers manually with a curl command

        $> curl -F "file=@somefile.tar.gz;type=application/vnd.redhat.<service-name>.somefile+tgz" -H "x-rh-identity: <base64 string>" -H "x-rh-insights-request-id: <uuid>" \
        http://localhost:3000/api/ingress/v1/upload

Note, that your service name needs to be in the `INGRESS_VALID_UPLOAD_TYPES` variable inside of the `.env` file.

For testing, the following base64 identity can be used:

    eyJpZGVudGl0eSI6IHsidHlwZSI6ICJVc2VyIiwgImFjY291bnRfbnVtYmVyIjogIjAwMDAwMDEiLCAib3JnX2lkIjogIjAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=

This decodes to:

    {"identity": {"type": "User", "account_number": "0000001", "org_id": "000001", "internal": {"org_id": "000001"}}}

#### Testing

Use `go test` to test the application

    $> make test
