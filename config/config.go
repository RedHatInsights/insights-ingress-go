package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// IngressConfig represents the runtime configuration
type IngressConfig struct {
	MaxSize              int64
	StageBucket          string
	RejectBucket         string
	Auth                 bool
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaAvailableTopic  string
	KafkaValidationTopic string
	KafkaTrackerTopic    string
	ValidTopics          []string
	Port                 int
	Profile              bool
	OpenshiftBuildCommit string
	Version              string
	InventoryURL         string
	MinioDev             bool
	MinioEndpoint        string
	MinioAccessKey       string
	MinioSecretKey       string
	Debug                bool
	DebugUserAgent       *regexp.Regexp
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
	options.SetDefault("InventoryURL", "http://inventory:8080/api/inventory/v1/hosts")
	options.SetDefault("Profile", false)
	options.SetDefault("Debug", false)
	options.SetDefault("DebugUserAgent", `unspecified`)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()
	commit := viper.New()
	commit.SetDefault("Openshift_Build_Commit", "notrunninginopenshift")
	commit.AutomaticEnv()

	return &IngressConfig{
		MaxSize:              options.GetInt64("MaxSize"),
		StageBucket:          options.GetString("StageBucket"),
		RejectBucket:         options.GetString("RejectBucket"),
		Auth:                 options.GetBool("Auth"),
		KafkaBrokers:         options.GetStringSlice("KafkaBrokers"),
		KafkaGroupID:         options.GetString("KafkaGroupID"),
		KafkaAvailableTopic:  options.GetString("KafkaAvailableTopic"),
		KafkaValidationTopic: options.GetString("KafkaValidationTopic"),
		KafkaTrackerTopic:    options.GetString("KafkaTrackerTopic"),
		ValidTopics:          strings.Split(options.GetString("ValidTopics"), ","),
		Port:                 options.GetInt("Port"),
		Profile:              options.GetBool("Profile"),
		Debug:                options.GetBool("Debug"),
		DebugUserAgent:       regexp.MustCompile(options.GetString("DebugUserAgent")),
		OpenshiftBuildCommit: commit.GetString("Openshift_Build_Commit"),
		Version:              "1.0.5",
		InventoryURL:         options.GetString("InventoryURL"),
		MinioDev:             options.GetBool("MinioDev"),
		MinioEndpoint:        options.GetString("MinioEndpoint"),
		MinioAccessKey:       options.GetString("MinioAccessKey"),
		MinioSecretKey:       options.GetString("MinioSecretKey"),
	}
}
