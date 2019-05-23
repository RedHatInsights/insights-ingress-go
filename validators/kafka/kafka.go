package kafka

import (
	"encoding/json"
	"log"

	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

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
	go queue.Consumer(kv.ValidationConsumerChannel, &queue.ConsumerConfig{
		Brokers: kv.KafkaBrokers,
		GroupID: kv.KafkaGroupID,
		Topic:   cfg.ValidationTopic,
	})

	go func() {
		for {
			data := <-kv.ValidationConsumerChannel
			ev := &validators.Response{}
			err := json.Unmarshal(data, ev)
			if err != nil {
				log.Printf("failed to unmarshal data: %v", err)
			} else {
				kv.RouteResponse(ev)
			}
		}
	}()

	return kv
}

// RouteResponse passes along responses based on their validation status
func (kv *Validator) RouteResponse(response *validators.Response) {
	switch response.Validation {
	case "success":
		kv.ValidChan <- response
	case "failure":
		kv.InvalidChan <- response
	default:
		log.Printf("Invalid validation in response: %s", response)
	}
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		log.Printf("failed to marshal json: %v", err)
		return
	}
	log.Printf("About to pass %s to testareno", data)
	kv.ValidationProducerMapping["platform.upload.testareno"] <- data
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan []byte)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers: kv.KafkaBrokers,
		Topic:   topic,
	})
	kv.ValidationProducerMapping[topic] = ch
}
