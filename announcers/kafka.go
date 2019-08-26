package announcers

import (
	"encoding/json"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"go.uber.org/zap"
)

// Kafka is an announcer that broadcases on a kafka topic
type Kafka struct {
	In chan []byte
}

// NewStatusAnnouncer creates a new announcer and starts the producer
func NewStatusAnnouncer(cfg *queue.ProducerConfig) *Kafka {
	k := &Kafka{
		In: make(chan []byte, 1000),
	}
	go queue.Producer(k.In, cfg)
	return k
}

// Status sends messages to the payload-tracker topic
func (k *Kafka) Status(vs *Status) {
	n := time.Now()
	vs.Service = "ingress"
	vs.Date = n.UTC()
	data, err := json.Marshal(vs)
	if err != nil {
		l.Log.Error("failed to marshal status message", zap.Error(err), zap.String("request_id", vs.RequestID))
		return
	}
	defer func() {
		l.Log.Debug("status announce", zap.Duration("duration", time.Since(n)))
	}()
	k.In <- data
}

// Stop the kafka input channel
func (k *Kafka) Stop() {
	close(k.In)
}
