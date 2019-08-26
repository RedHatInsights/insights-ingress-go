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
	tdMapping["unit2"] = "unit"
	tdMapping["openshift"] = "buckit"
}

// New constructs and initializes a new Kafka Validator
func New(cfg *Config, topics ...string) *Validator {
	kv := &Validator{
		ValidationProducerMapping: make(map[string]chan []byte),
		KafkaBrokers:              cfg.Brokers,
		KafkaGroupID:              cfg.GroupID,
	}
	for _, topic := range topics {
		topic = fmt.Sprintf("platform.upload.%s", topic)
		kv.addProducer(topic)
	}

	return kv
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.Error("failed to marshal json", zap.Error(err))
		return
	}
	topic := serviceToTopic(vr.Service)
	topic = fmt.Sprintf("platform.upload.%s", topic)
	l.Log.Debug("Posting data to topic", zap.ByteString("data", data), zap.String("topic", topic))
	kv.ValidationProducerMapping[topic] <- data
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan []byte, 100)
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
	return fmt.Sprintf("%s", service)
}
