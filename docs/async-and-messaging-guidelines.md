# Async and Messaging

## Architecture Overview

This service is a produce-only Kafka application. It does not consume messages. Two independent Kafka producers exist, each backed by a dedicated goroutine running `queue.Producer`:

1. **Validation announcer** -- publishes upload metadata to `platform.upload.announce` so downstream services can validate payloads.
2. **Status tracker** -- publishes status events to `platform.payload-status` for the payload-tracker service.

## Kafka Client Library

- Use `github.com/confluentinc/confluent-kafka-go/v2/kafka` (the confluent-kafka-go v2 module). Do not introduce alternative Kafka libraries.

## Producer Pattern

- Kafka producing flows through `internal/queue/queue.go` via the `Producer` function. Do not create additional producer implementations or call `kafka.NewProducer` outside this package.
- `Producer` reads from a `chan validators.ValidationMessage` and spawns a goroutine per message. Each goroutine creates its own `delivery_chan` to confirm delivery synchronously before returning.
- On delivery failure, `Producer` re-enqueues the message back onto the input channel (`in <- v`). Preserve this retry-by-requeue behavior when modifying producer logic.
- `kafka.PartitionAny` is used for all messages. Do not assign explicit partitions.
- When `Metadata.QueueKey` is present in the upload request, set `ValidationMessage.Key` to that value so related messages land on the same partition. See `internal/validators/kafka/kafka.go`.

## Channel Buffer Sizes

- The validation producer channel in `internal/validators/kafka/kafka.go` is buffered at 100: `make(chan validators.ValidationMessage, 100)`.
- The status announcer channel in `internal/announcers/kafka.go` is buffered at 1000: `make(chan validators.ValidationMessage, 1000)`.
- Preserve these buffer sizes or justify changes with load analysis.

## Topic Configuration

- Topic names are configured via `internal/config/config.go` in the `KafkaCfg` struct:
  - `KafkaAnnounceTopic` (default: `platform.upload.announce`)
  - `KafkaTrackerTopic` (default: `platform.payload-status`)
- When Clowder is enabled, topic names are resolved through `clowder.KafkaTopics[requestedName].Name`. Use `config.GetTopic()` for any topic name resolution at runtime.
- Do not hardcode topic name strings outside of config defaults.

## Message Schemas

- **Validation messages** (`validators.Request`): JSON-serialized struct containing `account`, `org_id`, `request_id`, `service`, `category`, `url`, `b64_identity`, `size`, `content_type`, `timestamp`, and nested `metadata`. Defined in `internal/validators/types.go`.
- **Status messages** (`announcers.Status`): JSON-serialized struct containing `service` (hardcoded to `"ingress"`), `account`, `org_id`, `request_id`, `status`, `status_msg`, `date`. Defined in `internal/announcers/kafka.go`.
- Validation messages carry a `Headers` map; the `service` header is set to `vr.Service` to allow topic-level filtering by downstream consumers.

## The Announcer Interface

- `internal/announcers/kafka.go` defines the `Announcer` interface with `Status(*Status)` and `Stop()` methods.
- Use `announcers.NewStatusAnnouncer(*queue.ProducerConfig)` to create the Kafka-backed implementation.
- Call `Stop()` to close the input channel and shut down the producer.

## The Validator Interface

- `internal/validators/types.go` defines the `Validator` interface with `Validate(*Request)` and `ValidateService(*ServiceDescriptor) error`.
- `internal/validators/kafka/kafka.go` implements this interface. `New()` accepts a `*Config` and a variadic list of valid service names.
- `ValidateService` checks against a `map[string]bool` built from `ValidUploadTypes` config. This check happens before any Kafka message is produced.

## Kafka Security Configuration

- SSL/SASL settings flow from `config.KafkaSSLCfg` through `queue.ProducerConfig` to `kafka.ConfigMap`:
  - Set `ssl.ca.location` only when `CA` is non-empty.
  - Set `security.protocol`, `sasl.mechanism`, `sasl.username`, `sasl.password` only when `SASLMechanism` is non-empty.
- In Clowder environments, credentials come from `broker.Sasl` and CA from `broker.Cacert`.

## Testing Fakes

- Use `validators.Fake` (in `internal/validators/fake.go`) when testing upload handlers without Kafka.
- Use `announcers.Fake` (in `internal/announcers/fake.go`) when testing without the status announcer.
- Prefer these fakes over mocking the Kafka client directly.

## Message Flow Summary

1. HTTP upload arrives at `internal/upload/upload.go` handler.
2. Handler sends `"received"` status via `tracker.Status()` (produces to `platform.payload-status`).
3. Payload is staged to S3/MinIO/filesystem.
4. Handler sends `"success"` status via `tracker.Status()`.
5. Handler calls `validator.Validate(vr)` which produces the `validators.Request` JSON to `platform.upload.announce` with the `service` header.
6. HTTP response is returned (202 Accepted, or 201 Created for advisor without metadata).

## Verification

```bash
# Confirm only confluent-kafka-go is used as Kafka library
grep -r "kafka" go.mod | grep -v "confluent-kafka-go"

# Confirm Kafka producing goes through queue.Producer
grep -rn "kafka.NewProducer" internal/ --include="*.go"

# Confirm channel buffer sizes
grep -n "make(chan validators.ValidationMessage" internal/ -r --include="*.go"

# Confirm topic names are not hardcoded outside config
grep -rn '"platform\.' internal/ --include="*.go" | grep -v config.go | grep -v _test.go

# Run tests
go test ./internal/validators/kafka/... ./internal/upload/... ./internal/announcers/...
```
