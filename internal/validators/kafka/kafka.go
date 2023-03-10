package kafka

import (
	"encoding/json"
	"errors"

	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/queue"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	"github.com/sirupsen/logrus"
)

var tdMapping map[string]string

func init() {
	tdMapping = make(map[string]string)
	// Map these service names (the key) to the topics (the value)
	tdMapping["unit2"] = "unit"
	tdMapping["openshift"] = "buckit"
}

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
	validServicesMap          map[string]bool
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
func New(cfg *Config, topics ...string) *Validator {
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

	kv.validServicesMap = buildValidServicesMap(topics)

	topic := config.GetTopic("platform.upload.announce")

	kv.addProducer(topic)

	return kv
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.WithFields(logrus.Fields{"error": err}).Error("failed to marshal json")
		return
	}
	announceTopic := config.Get().KafkaConfig.KafkaAnnounceTopic
	l.Log.WithFields(logrus.Fields{"data": data, "topic": announceTopic}).Debug("Posting data to topic")
	message := validators.ValidationMessage{
		Message: data,
		Headers: map[string]string{
			"service": vr.Service,
		},
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
	topic := serviceToTopic(service.Service)

	_, isValidService := kv.validServicesMap[topic]

	if isValidService {
		return nil
	}

	return errors.New("Validation topic is invalid: " + topic)
}

func serviceToTopic(service string) string {
	topic := tdMapping[service]
	if topic != "" {
		return topic
	}
	return service
}

func buildValidServicesMap(validServicesList []string) map[string]bool {

	validServicesMap := make(map[string]bool)

	for _, service := range validServicesList {
		validServicesMap[service] = true
	}

	return validServicesMap
}
