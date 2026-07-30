package kafka

import (
	"context"
	"encoding/json"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/queue"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	"github.com/sirupsen/logrus"
)

// Validator posts requests to topics for validation
type Validator struct {
	ValidationProducerChannel chan validators.ValidationMessage
	KafkaBrokers              []string
	KafkaGroupID              string
	Username                  string
	Password                  string
	CA                        string
	SASLMechanism             string
	KafkaSecurityProtocol     string
	announceTopic             string
	validUploadTypes          map[string]bool
}

// Config configures a new Kafka Validator
type Config struct {
	Brokers               []string
	GroupID               string
	ValidationTopic       string
	Username              string
	Password              string
	CA                    string
	KafkaSecurityProtocol string
	SASLMechanism         string
	Debug                 bool
}

// New constructs and initializes a new Kafka Validator
func New(cfg *Config, validServices ...string) *Validator {
	kv := &Validator{
		ValidationProducerChannel: make(chan validators.ValidationMessage),
		KafkaBrokers:              cfg.Brokers,
		KafkaGroupID:              cfg.GroupID,
		KafkaSecurityProtocol:     cfg.KafkaSecurityProtocol,
	}

	if cfg.CA != "" {
		kv.CA = cfg.CA
	}

	if cfg.Username != "" {
		kv.Username = cfg.Username
		kv.Password = cfg.Password
	}

	if cfg.SASLMechanism != "" {
		kv.SASLMechanism = cfg.SASLMechanism
	}

	kv.validUploadTypes = buildValidUploadTypeMap(validServices)
	kv.announceTopic = config.Get().KafkaConfig.KafkaAnnounceTopic

	kv.addProducer(kv.announceTopic)

	return kv
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(ctx context.Context, vr *validators.Request) {
	ctx, span := otel.Tracer("ingress").Start(ctx, "send "+kv.announceTopic,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.operation.name", "send"),
			attribute.String("messaging.operation.type", "send"),
			attribute.String("messaging.destination.name", kv.announceTopic),
			attribute.String("rh.content_type", vr.Service),
			attribute.Int64("rh.payload_size", vr.Size),
		))
	defer span.End()

	data, err := json.Marshal(vr)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal validation request")
		l.Log.WithFields(logrus.Fields{"error": err}).Error("failed to marshal json")
		return
	}
	l.Log.WithFields(logrus.Fields{"data": data, "topic": kv.announceTopic}).Debug("Posting data to topic")
	headers := map[string]string{
		"service": vr.Service,
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	message := validators.ValidationMessage{
		Message: data,
		Headers: headers,
	}
	if vr.Metadata.QueueKey != "" {
		message.Key = []byte(vr.Metadata.QueueKey)
	}

	kv.ValidationProducerChannel <- message
	incMessageProduced(vr.Service)
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan validators.ValidationMessage, 100)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers:               kv.KafkaBrokers,
		Topic:                 topic,
		CA:                    kv.CA,
		Username:              kv.Username,
		Password:              kv.Password,
		KafkaSecurityProtocol: kv.KafkaSecurityProtocol,
		SASLMechanism:         kv.SASLMechanism,
	})
	kv.ValidationProducerChannel = ch
}

// ValidateService ensures that a service maps to a real topic
func (kv *Validator) ValidateService(service *validators.ServiceDescriptor) error {

	_, isValidUploadType := kv.validUploadTypes[service.Service]

	if isValidUploadType {
		return nil
	}

	return errors.New("Upload type is not supported: " + service.Service)
}

func buildValidUploadTypeMap(validUploadTypeList []string) map[string]bool {

	validUploadTypes := make(map[string]bool)

	for _, service := range validUploadTypeList {
		validUploadTypes[service] = true
	}

	return validUploadTypes
}
