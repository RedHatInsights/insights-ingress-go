package queue

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

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
		w.WriteMessages(context.Background(),
			kafka.Message{
				Key:   nil,
				Value: v,
			},
		)
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
