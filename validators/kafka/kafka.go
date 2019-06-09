package kafka

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

var tdMapping map[string]string

func init() {
	tdMapping = make(map[string]string)
	tdMapping["openshift"] = "platform.upload.buckit"
}

// New constructs and initializes a new Kafka Validator
func New(cfg *Config, topics ...string) *Validator {
	kv := &Validator{
		ValidationProducerMapping: make(map[string]chan []byte),
		ValidationConsumerChannel: make(chan []byte),
		KafkaBrokers:              cfg.Brokers,
		KafkaGroupID:              cfg.GroupID,
		ValidChan:                 cfg.ValidChan,
		InvalidChan:               cfg.InvalidChan,
	}
	for _, topic := range topics {
		kv.addProducer(topic)
	}
	go queue.Consumer(cfg.Context, kv.ValidationConsumerChannel, &queue.ConsumerConfig{
		Brokers: kv.KafkaBrokers,
		GroupID: kv.KafkaGroupID,
		Topic:   cfg.ValidationTopic,
	})

	go func() {
		for data := range kv.ValidationConsumerChannel {
			ev := &validators.Response{}
			err := json.Unmarshal(data, ev)
			if err != nil {
				l.Log.Error("failed to unmarshal data", zap.Error(err))
			} else {
				kv.RouteResponse(ev)
			}
		}
		l.Log.Info("consumer channel closed, shutting down")
		close(kv.ValidChan)
		close(kv.InvalidChan)
	}()

	return kv
}

// RouteResponse passes along responses based on their validation status
func (kv *Validator) RouteResponse(response *validators.Response) {
	// Since we only want to track the elapsed times of responses with
	// Timestamps, we need a second counter to make we at least count the
	// number of responses accurately
	if !response.Timestamp.IsZero() {
		observeValidationElapsed(response.Timestamp, response.Validation)
	}
	inc(response.Validation)
	switch response.Validation {
	case "success":
		kv.ValidChan <- response
	case "failure":
		kv.InvalidChan <- response
	default:
		l.Log.Error("Invalid validation in response", zap.String("response.validation", response.Validation))
		return
	}
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.Error("failed to marshal json", zap.Error(err))
		return
	}
	topic := serviceToTopic(vr.Service)
	l.Log.Debug("Posting data to topic", zap.ByteString("data", data), zap.String("topic", topic))
	kv.ValidationProducerMapping[topic] <- data
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan []byte)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers: kv.KafkaBrokers,
		Topic:   topic,
	})
	kv.ValidationProducerMapping[topic] = ch
}

// ValidateService ensures that a service maps to a real topic
func (kv *Validator) ValidateService(service *validators.ServiceDescriptor) error {
	topic := serviceToTopic(service.Service)
	for _, validTopic := range config.Get().ValidTopics {
		if validTopic == topic {
			return nil
		}
	}
	return errors.New("Validation topic is invalid")
}

func serviceToTopic(service string) string {
	topic := tdMapping[service]
	if topic != "" {
		return topic
	}
	return fmt.Sprintf("platform.upload.%s", service)
}
