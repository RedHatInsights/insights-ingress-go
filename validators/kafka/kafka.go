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
				if ev.Validation == "success" {
					cfg.ValidChan <- ev
				} else if ev.Validation == "failure" {
					cfg.InvalidChan <- ev
				}
			}
		}
	}()

	return kv
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
<<<<<<< HEAD
=======

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
		log.Printf("got %s about to write to kafka", v)
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
>>>>>>> Log the json as a string
