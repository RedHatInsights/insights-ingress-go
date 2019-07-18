package announcers

import (
	"encoding/json"
	"time"

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

// NewStatusAnnouncer creates a new announcer and starts the producer
func NewStatusAnnouncer(cfg *queue.ProducerConfig) *Kafka {
	k := &Kafka{
		In: make(chan []byte, 1000),
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
	n := time.Now()
	defer func() {
		l.Log.Info("announce", zap.Duration("duration", time.Since(n)))
	}()
	k.In <- data
}

// Status sends messages to the payload-tracker topic
func (k *Kafka) Status(vs *validators.Status) {
	data, err := json.Marshal(vs)
	if err != nil {
		l.Log.Error("failed to marshal status message", zap.Error(err), zap.String("request_id", vs.RequestID))
		return
	}
	n := time.Now()
	defer func() {
		l.Log.Info("status announce", zap.Duration("duration", time.Since(n)))
	}()
	k.In <- data
}

// Stop the kafka input channel
func (k *Kafka) Stop() {
	close(k.In)
}
