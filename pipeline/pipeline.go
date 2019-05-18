package pipeline

import (
	"encoding/json"
	"log"

	"cloud.redhat.com/ingress/config"
	"cloud.redhat.com/ingress/queue"
	"cloud.redhat.com/ingress/stage"
)

// NewKafkaValidator constructs and initializes a new Kafka Validator
func NewKafkaValidator(cfg *config.IngressConfig) *KafkaValidator {
	kv := &KafkaValidator{
		ValidationProducerMapping: make(map[string]chan []byte),
		ValidationConsumerChannel: make(chan []byte),
		AvailableProducerChannel:  make(chan []byte),
		KafkaBrokers:              cfg.KafkaBrokers,
		KafkaGroupID:              cfg.KafkaGroupID,
	}
	kv.addProducer(cfg.KafkaAvailableTopic)
	kv.addProducer("platform.upload.testareno")
	return kv
}

// Validate validates a ValidationRequest
func (kv *KafkaValidator) Validate(vr *ValidationRequest) {
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

// Submit accepts a stage request and a validation request
func (p *Pipeline) Submit(in *stage.Input, vr *ValidationRequest) {
	url, err := p.Stager.Stage(in)
	if err != nil {
		log.Printf("Error staging %v: %v", in, err)
		return
	}
	vr.URL = url
	p.Validator.Validate(vr)
}

func (p *Pipeline) run() {
	for {

	}
}
