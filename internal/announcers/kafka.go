package announcers

import (
	"encoding/json"
	"time"

	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
	"github.com/redhatinsights/insights-ingress-go/internal/queue"
	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	"github.com/sirupsen/logrus"
)

// Kafka is an announcer that broadcases on a kafka topic
type Kafka struct {
	In chan validators.ValidationMessage
}

type Announcer interface {
	Status(e *Status)
	Stop()
}

// Status is the message sent to the payload tracker
type Status struct {
	Service     string    `json:"service"`
	Source      string    `json:"source,omitempty"`
	Account     string    `json:"account"`
	RequestID   string    `json:"request_id"`
	InventoryID string    `json:"inventory_id,omitempty"`
	SystemID    string    `json:"system_id,omitempty"`
	Status      string    `json:"status"`
	StatusMsg   string    `json:"status_msg"`
	Date        time.Time `json:"date"`
}

// NewStatusAnnouncer creates a new announcer and starts the producer
func NewStatusAnnouncer(cfg *queue.ProducerConfig) *Kafka {
	k := &Kafka{
		In: make(chan validators.ValidationMessage, 1000),
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
		l.Log.WithFields(logrus.Fields{"request_id": vs.RequestID, "error": err}).Error("Failed to marshal status message")
		return
	}
	defer func() {
		l.Log.WithFields(logrus.Fields{"duration": time.Since(n)}).Debug("status announce")
	}()
	message := validators.ValidationMessage{
		Message: data,
	}
	k.In <- message
}

// Stop the kafka input channel
func (k *Kafka) Stop() {
	close(k.In)
}
