package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaHealthChecker verifies that Kafka accepts a metadata request.
type KafkaHealthChecker struct {
	admin   *kafka.AdminClient
	timeout time.Duration
	err     error
}

// NewKafkaHealthChecker creates a Kafka metadata checker from the producer's
// connection settings. It does not produce messages.
func NewKafkaHealthChecker(config ProducerConfig, timeout time.Duration) *KafkaHealthChecker {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	configMap := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(config.Brokers, ","),
	}
	if config.KafkaSecurityProtocol != "" {
		_ = configMap.SetKey("security.protocol", config.KafkaSecurityProtocol)
	}
	if config.CA != "" {
		_ = configMap.SetKey("ssl.ca.location", config.CA)
	}
	if config.SASLMechanism != "" {
		_ = configMap.SetKey("sasl.mechanism", config.SASLMechanism)
		_ = configMap.SetKey("sasl.username", config.Username)
		_ = configMap.SetKey("sasl.password", config.Password)
	}
	admin, err := kafka.NewAdminClient(&configMap)
	return &KafkaHealthChecker{admin: admin, timeout: timeout, err: err}
}

// Check verifies broker metadata access.
func (c *KafkaHealthChecker) Check(ctx context.Context) error {
	if c.err != nil {
		return fmt.Errorf("kafka admin client unavailable: %w", c.err)
	}
	if c.admin == nil {
		return errors.New("kafka admin client unavailable")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := c.admin.GetMetadata(nil, false, int(c.timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("kafka metadata unavailable: %w", err)
	}
	return nil
}

// Close releases the Kafka client resources used by the health checker.
func (c *KafkaHealthChecker) Close() {
	if c.admin != nil {
		c.admin.Close()
	}
}
