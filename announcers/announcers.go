package announcers

import (
	"encoding/json"
	"log"

	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
)

// Fake is a fake announcer
type Fake struct {
	event *validators.Response
}

// Announce does nothing
func (f *Fake) Announce(e *validators.Response) {
	f.event = e
}

// Kafka is an announcer that broadcases on a kafka topic
type Kafka struct {
	In chan []byte
}

// NewKafkaAnnouncer creates a new announcer and starts the associated producer
func NewKafkaAnnouncer(cfg *queue.ProducerConfig) *Kafka {
	k := &Kafka{
		In: make(chan []byte),
	}
	go queue.Producer(k.In, cfg)
	return k
}

// Announce broadcasts the response
func (k *Kafka) Announce(e *validators.Response) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("failed to marshal json: %v", err)
		return
	}
	k.In <- data
}
