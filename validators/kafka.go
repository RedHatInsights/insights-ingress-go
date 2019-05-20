package validators

import (
	"encoding/json"
	"log"

	"cloud.redhat.com/ingress/queue"
	"cloud.redhat.com/ingress/announcers"
)

type KafkaConfig struct {
	Brokers []string
	GroupID string
	AvailableTopic string
	ValidationTopic string
	AnnouncerChan chan *announcers.AvailableEvent
}

// NewKafkaValidator constructs and initializes a new Kafka Validator
func NewKafkaValidator(cfg *KafkaConfig, topics ...string) *KafkaValidator {
	kv := &KafkaValidator{
		ValidationProducerMapping: make(map[string]chan []byte),
		ValidationConsumerChannel: make(chan []byte),
		KafkaBrokers:              cfg.Brokers,
		KafkaGroupID:              cfg.GroupID,
	}
	kv.addProducer(cfg.AvailableTopic)
	for _, topic := range topics {
		kv.addProducer(topic)
	}
	go queue.Consumer(kv.ValidationConsumerChannel, &queue.ConsumerConfig{
		Brokers: kv.KafkaBrokers,
		GroupID: kv.KafkaGroupID,
		Topic: cfg.ValidationTopic,
	})

	go func() {
		for {
			data := <- kv.ValidationConsumerChannel
			ev := &announcers.AvailableEvent{}
			err := json.Unmarshal(data, ev)
			if err != nil {
				log.Printf("failed to unarshal data: %v", err)
			} else {
				cfg.AnnouncerChan <- ev
			}
		}
	}()

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