package config

import (
	"strings"

	"time"

	"github.com/spf13/viper"
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
	KafkaTrackerTopic           string
	ValidTopics                 []string
	Port                        int
	Profile                     bool
	OpenshiftBuildCommit        string
	Version                     string
	Simulate                    bool
	SimulationStageDelay        time.Duration
	SimulationValidateCallDelay time.Duration
	SimulationValidateDelay     time.Duration
	InventoryURL                string
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
	options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	options.SetDefault("ValidTopics", "unit")
	options.SetDefault("OpenshiftBuildCommit", "notrunninginopenshift")
	options.SetDefault("Simulate", false)
	options.SetDefault("SimulationStageDelay", 100)
	options.SetDefault("SimulationValidateDelay", 5000)
	options.SetDefault("SimulationValidateCallDelay", 100)
	options.SetDefault("InventoryURL", "http://inventory:8080/api/inventory/v1/hosts")
	options.SetDefault("Profile", false)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()
	commit := viper.New()
	commit.SetDefault("Openshift_Build_Commit", "notrunninginopenshift")
	commit.AutomaticEnv()

	return &IngressConfig{
		MaxSize:                     options.GetInt("MaxSize"),
		StageBucket:                 options.GetString("StageBucket"),
		RejectBucket:                options.GetString("RejectBucket"),
		Auth:                        options.GetBool("Auth"),
		KafkaBrokers:                options.GetStringSlice("KafkaBrokers"),
		KafkaGroupID:                options.GetString("KafkaGroupID"),
		KafkaAvailableTopic:         options.GetString("KafkaAvailableTopic"),
		KafkaValidationTopic:        options.GetString("KafkaValidationTopic"),
		KafkaTrackerTopic:           options.GetString("KafkaTrackerTopic"),
		ValidTopics:                 strings.Split(options.GetString("ValidTopics"), ","),
		Port:                        options.GetInt("Port"),
		Profile:                     options.GetBool("Profile"),
		OpenshiftBuildCommit:        commit.GetString("Openshift_Build_Commit"),
		Version:                     "1.0.2",
		Simulate:                    options.GetBool("Simulate"),
		SimulationStageDelay:        options.GetDuration("SimulationStageDelay"),
		SimulationValidateCallDelay: options.GetDuration("SimulationValidateCallDelay"),
		SimulationValidateDelay:     options.GetDuration("SimulationValidateDelay"),
		InventoryURL:                options.GetString("InventoryURL"),
	}
}
