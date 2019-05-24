package kafka

import (
	"encoding/json"

	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/insights-ingress-go/queue"
	"github.com/redhatinsights/insights-ingress-go/validators"
	"go.uber.org/zap"
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
				l.Log.Error("failed to unmarshal data", zap.Error(err))
			} else {
				kv.RouteResponse(ev)
			}
		}
	}()

	return kv
}

// RouteResponse passes along responses based on their validation status
func (kv *Validator) RouteResponse(response *validators.Response) {
	inc(response.Validation)
	switch response.Validation {
	case "success":
		kv.ValidChan <- response
	case "failure":
		kv.InvalidChan <- response
	default:
		l.Log.Error("Invalid validation in response", zap.String("response.validation", response.Validation))
	}
}

// Validate validates a ValidationRequest
func (kv *Validator) Validate(vr *validators.Request) {
	data, err := json.Marshal(vr)
	if err != nil {
		l.Log.Error("failed to marshal json", zap.Error(err))
		return
	}
	topic := "platform.upload.testareno"
	l.Log.Debug("Posting data to topic", zap.ByteString("data", data), zap.String("topic", topic))
	kv.ValidationProducerMapping[topic] <- data
}

func (kv *Validator) addProducer(topic string) {
	ch := make(chan []byte)
	go queue.Producer(ch, &queue.ProducerConfig{
		Brokers: kv.KafkaBrokers,
		Topic:   topic,
	})
	kv.ValidationProducerMapping[topic] = ch
}
