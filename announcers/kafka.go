package announcers

import (
	"encoding/json"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
)

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
func (k *Kafka) Announce(vr *validators.Response) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.Error("failed to marshal json", zap.Error(err), zap.String("request_id", vr.RequestID))
		return
	}
	k.In <- data
}

func (k *Kafka) Stop() {
	close(k.In)
}
