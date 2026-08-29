package health

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaConfig contains the Kafka connection settings needed for a broker
// metadata check. No messages are produced by this check.
type KafkaConfig struct {
	Brokers          []string
	SecurityProtocol string
	CA               string
	Username         string
	Password         string
	SASLMechanism    string
	Timeout          time.Duration
}

// KafkaChecker verifies that Kafka accepts a metadata request.
type KafkaChecker struct {
	admin   *kafka.AdminClient
	timeout time.Duration
	err     error
}

func NewKafkaChecker(cfg KafkaConfig) *KafkaChecker {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 2 * time.Second
	}
	configMap := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
	}
	if cfg.SecurityProtocol != "" {
		_ = configMap.SetKey("security.protocol", cfg.SecurityProtocol)
	}
	if cfg.CA != "" {
		_ = configMap.SetKey("ssl.ca.location", cfg.CA)
	}
	if cfg.SASLMechanism != "" {
		_ = configMap.SetKey("sasl.mechanism", cfg.SASLMechanism)
		_ = configMap.SetKey("sasl.username", cfg.Username)
		_ = configMap.SetKey("sasl.password", cfg.Password)
	}
	admin, err := kafka.NewAdminClient(&configMap)
	return &KafkaChecker{admin: admin, timeout: cfg.Timeout, err: err}
}

func (c *KafkaChecker) Check(ctx context.Context) error {
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

func (c *KafkaChecker) Close() {
	if c.admin != nil {
		c.admin.Close()
	}
}
