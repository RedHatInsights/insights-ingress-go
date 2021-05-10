package kafka

import (
	"encoding/json"
	"errors"
	"fmt"

	rhiconfig "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"github.com/sirupsen/logrus"
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

	if cfg.CA != "" {
		kv.CA = cfg.CA
	}

	if cfg.Username != "" {
		kv.Username = cfg.Username
		kv.Password = cfg.Password
	}

	for _, topic := range topics {
		var realizedTopicName string
		topic = fmt.Sprintf("platform.upload.%s", topic)
		if config.Get().UseClowder {
			realizedTopicName = rhiconfig.KafkaTopics[topic].Name
		} else {
			realizedTopicName = topic
		}
		kv.addProducer(realizedTopicName)
	}

	return kv
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("failed to marshal json")
		return
	}
	topic := serviceToTopic(vr.Service)
	topic = fmt.Sprintf("platform.upload.%s", topic)
	var realizedTopicName string
	if config.Get().UseClowder {
		realizedTopicName = rhiconfig.KafkaTopics[topic].Name
	} else {
		realizedTopicName = topic
	}
	l.Log.WithFields(logrus.Fields{"data": data, "topic": realizedTopicName}).Debug("Posting data to topic")
	kv.ValidationProducerMapping[realizedTopicName] <- data
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan []byte, 100)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers:  kv.KafkaBrokers,
		Topic:    topic,
		CA:       kv.CA,
		Username: kv.Username,
		Password: kv.Password,
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
