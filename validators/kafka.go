package validators

import (
	"encoding/json"
	"log"

	"cloud.redhat.com/ingress/config"
	"cloud.redhat.com/ingress/queue"
)

// NewKafkaValidator constructs and initializes a new Kafka Validator
func NewKafkaValidator(cfg *config.IngressConfig, topics ...string) *KafkaValidator {
	kv := &KafkaValidator{
		ValidationProducerMapping: make(map[string]chan []byte),
		ValidationConsumerChannel: make(chan []byte),
		KafkaBrokers:              cfg.KafkaBrokers,
		KafkaGroupID:              cfg.KafkaGroupID,
	}
	kv.addProducer(cfg.KafkaAvailableTopic)
	for _, topic := range topics {
		kv.addProducer(topic)
	}
	return kv
}

// Validate validates a ValidationRequest
func (kv *KafkaValidator) Validate(vr *Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		log.Printf("failed to marshal json: %v", err)
		return
	}
	log.Printf("About to pass %v to testareno", data)
	kv.ValidationProducerMapping["platform.upload.testareno"] <- data
}

func (kv *KafkaValidator) addProducer(topic string) {
	ch := make(chan []byte)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers: kv.KafkaBrokers,
		Topic:   topic,
	})
	kv.ValidationProducerMapping[topic] = ch
}
