package config

import (
	"strings"

	"github.com/spf13/viper"
	"time"
)

// IngressConfig represents the runtime configuration
type IngressConfig struct {
	MaxSize                     int
	StageBucket                 string
	RejectBucket                string
	Auth                        bool
	KafkaBrokers                []string
	KafkaGroupID                string
	KafkaAvailableTopic         string
	KafkaValidationTopic        string
	ValidTopics                 []string
	Port                        int
	Simulate                    bool
	SimulationStageDelay        time.Duration
	SimulationValidateCallDelay time.Duration
	SimulationValidateDelay     time.Duration
}

// Get returns an initialized IngressConfig
func Get() *IngressConfig {
	options := viper.New()
	options.SetDefault("MaxSize", 10*1024*1024)
	options.SetDefault("Port", 3000)
	options.SetDefault("StageBucket", "available")
	options.SetDefault("RejectBucket", "rejected")
	options.SetDefault("Auth", true)
	options.SetDefault("KafkaBrokers", []string{"kafka:29092"})
	options.SetDefault("KafkaGroupID", "ingress")
	options.SetDefault("KafkaAvailableTopic", "platform.upload.available")
	options.SetDefault("KafkaValidationTopic", "platform.upload.validation")
	options.SetDefault("ValidTopics", "unit")
	options.SetDefault("Simulate", false)
	options.SetDefault("SimulationStageDelay", 100)
	options.SetDefault("SimulationValidateDelay", 5000)
	options.SetDefault("SimulationValidateCallDelay", 100)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()

	return &IngressConfig{
		MaxSize:                     options.GetInt("MaxSize"),
		StageBucket:                 options.GetString("StageBucket"),
		RejectBucket:                options.GetString("RejectBucket"),
		Auth:                        options.GetBool("Auth"),
		KafkaBrokers:                options.GetStringSlice("KafkaBrokers"),
		KafkaGroupID:                options.GetString("KafkaGroupID"),
		KafkaAvailableTopic:         options.GetString("KafkaAvailableTopic"),
		KafkaValidationTopic:        options.GetString("KafkaValidationTopic"),
		ValidTopics:                 strings.Split(options.GetString("ValidTopics"), ","),
		Port:                        options.GetInt("Port"),
		Simulate:                    options.GetBool("Simulate"),
		SimulationStageDelay:        options.GetDuration("SimulationStageDelay"),
		SimulationValidateCallDelay: options.GetDuration("SimulationValidateCallDelay"),
		SimulationValidateDelay:     options.GetDuration("SimulationValidateDelay"),
	}
}
