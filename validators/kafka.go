package validators

import (
	"context"
	"encoding/json"
	"log"

	"cloud.redhat.com/ingress/announcers"
	"github.com/segmentio/kafka-go"
)


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
	go Consumer(kv.ValidationConsumerChannel, &ConsumerConfig{
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
	go Producer(ch, &ProducerConfig{
		Brokers: kv.KafkaBrokers,
		Topic:   topic,
	})
	kv.ValidationProducerMapping[topic] = ch
}



//Producer consumes in and produces to the topic in config
func Producer(in chan []byte, config *ProducerConfig) {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  config.Brokers,
		Topic:    config.Topic,
		Balancer: &kafka.Hash{},
	})

	defer w.Close()

	for {
		v := <-in
		log.Printf("got %v about to write to kafka", v)
		err := w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   nil,
				Value: v,
			},
		)
		if err != nil {
			log.Printf("error while writing: %v", err)
		}
	}
}

// Consumer consumes a topic and puts the messages into out
func Consumer(out chan []byte, config *ConsumerConfig) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		GroupID: config.GroupID,
		Topic:   config.Topic,
	})

	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message: %v", err)
		} else {
			out <- m.Value
		}
	}
}