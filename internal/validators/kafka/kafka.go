package kafka

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	rhiconfig "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/queue"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	"github.com/sirupsen/logrus"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

var tdMapping map[string]string

func init() {
	tdMapping = make(map[string]string)
	tdMapping["unit2"] = "unit"
	tdMapping["openshift"] = "buckit"
}

// Validator posts requests to topics for validation
type Validator struct {
	ValidationProducerMapping map[string]chan validators.ValidationMessage
	KafkaBrokers              []string
	KafkaGroupID              string
	Username                  string
	Password                  string
	CA                        string
	SASLMechanism             string
	Protocol                  string
	BufferResponse			  chan bool 
}

// Config configures a new Kafka Validator
type Config struct {
	Brokers         []string
	GroupID         string
	ValidationTopic string
	Username        string
	Password        string
	CA              string
	Protocol        string
	SASLMechanism   string
}

// New constructs and initializes a new Kafka Validator
func New(cfg *Config, topics ...string) *Validator {
	kv := &Validator{
		ValidationProducerMapping: make(map[string]chan validators.ValidationMessage),
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

	if cfg.SASLMechanism != "" {
		kv.SASLMechanism = cfg.SASLMechanism
		kv.Protocol = cfg.Protocol
	}

	// ensure the announce topic is added and valid
	topics = append(topics, "announce")

	for _, topic := range topics {
		var realizedTopicName string
		topic = fmt.Sprintf("platform.upload.%s", topic)
		if clowder.IsClowderEnabled() {
			realizedTopicName = rhiconfig.KafkaTopics[topic].Name
		} else {
			realizedTopicName = topic
		}
		kv.addProducer(realizedTopicName)
	}

	return kv
}

// Load messages puts the incoming message on the kafka topics
func (kv *Validator) Load(message validators.ValidationMessage, realizedTopicName string) <- chan bool {
	kv.ValidationProducerMapping[realizedTopicName] <- message
	kv.ValidationProducerMapping[config.Get().KafkaConfig.KafkaAnnounceTopic] <- message
	kv.BufferResponse <- true
	return kv.BufferResponse
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) bool {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("failed to marshal json")
		return false
	}
	topic := serviceToTopic(vr.Service)
	topic = fmt.Sprintf("platform.upload.%s", topic)
	var realizedTopicName string
	if clowder.IsClowderEnabled() {
		realizedTopicName = rhiconfig.KafkaTopics[topic].Name
	} else {
		realizedTopicName = topic
	}
	l.Log.WithFields(logrus.Fields{"data": data, "topic": realizedTopicName}).Debug("Posting data to topic")
	message := validators.ValidationMessage{
		Message: data,
		Headers: map[string]string{
			"service": vr.Service,
		},
	}
	// Load the message into the buffer and set a timeout. If the timeout finishes before we're loaded into the buffer,
	// then we can send the proper status code that our buffer is full.
	select {
	case <- kv.Load(message, realizedTopicName):
		l.Log.WithFields(logrus.Fields{"topic": realizedTopicName}).Debug("Message loaded to buffer for publishing")
		return true
	case <- time.After(time.Duration(config.Get().KafkaConfig.KafkaLoadWaitTime) * time.Second):
		l.Log.WithFields(logrus.Fields{"topic": realizedTopicName, "request_id": vr.RequestID}).Error("Buffer is full, waiting for space")
		return false
	}
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan validators.ValidationMessage, 100)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers:       kv.KafkaBrokers,
		Topic:         topic,
		CA:            kv.CA,
		Username:      kv.Username,
		Password:      kv.Password,
		Protocol:      kv.Protocol,
		SASLMechanism: kv.SASLMechanism,
	})
	kv.ValidationProducerMapping[topic] = ch
}

// ValidateService ensures that a service maps to a real topic
func (kv *Validator) ValidateService(service *validators.ServiceDescriptor) error {
	topic := serviceToTopic(service.Service)
	for _, validTopic := range config.Get().KafkaConfig.ValidTopics {
		if validTopic == topic {
			return nil
		}
	}
	return errors.New("Validation topic is invalid: " + topic)
}

func serviceToTopic(service string) string {
	topic := tdMapping[service]
	if topic != "" {
		return topic
	}
	return fmt.Sprintf("%s", service)
}
