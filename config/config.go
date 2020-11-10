package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"

	"github.com/spf13/viper"
)

// IngressConfig represents the runtime configuration
type IngressConfig struct {
	Hostname             string
	MaxSize              int64
	StageBucket          string
	Auth                 bool
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaTrackerTopic    string
	ValidTopics          []string
	WebPort              int
	MetricsPort          int
	Profile              bool
	OpenshiftBuildCommit string
	Version              string
	MinioEndpoint        string
	MinioAccessKey       string
	MinioSecretKey       string
	UseSSL               bool
	Debug                bool
	DebugUserAgent       *regexp.Regexp
	UseClowder           bool
}

// Get returns an initialized IngressConfig
func Get() *IngressConfig {

	options := viper.New()

	if os.Getenv("CLOWDER_ENABLED") == "true" {
		cfg := clowder.LoadedConfig

		sb := os.Getenv("INGRESS_STAGEBUCKET")
		bucket, _ := clowder.ObjectBuckets[sb]

		options.SetDefault("WebPort", cfg.WebPort)
		options.SetDefault("MetricsPort", cfg.MetricsPort)
		options.SetDefault("KafkaBrokers", fmt.Sprintf("%s:%v", cfg.Kafka.Brokers[0].Hostname, *cfg.Kafka.Brokers[0].Port))
		options.SetDefault("MinioEndpoint", fmt.Sprintf("%s:%d", cfg.ObjectStore.Hostname, cfg.ObjectStore.Port))
		options.SetDefault("MinioAccessKey", *cfg.ObjectStore.Buckets[0].AccessKey)
		options.SetDefault("MinioSecretKey", *cfg.ObjectStore.Buckets[0].SecretKey)
		options.SetDefault("UseSSL", cfg.ObjectStore.Tls)
		options.SetDefault("StageBucket", bucket.RequestedName)
	} else {
		options.SetDefault("WebPort", 3000)
		options.SetDefault("MetricsPort", 8080)
		options.SetDefault("KafkaBrokers", []string{"kafka:29092"})
		options.SetDefault("StageBucket", "available")
		options.SetDefault("UseSSL", false)
	}

	options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	options.SetDefault("KafkaGroupID", "ingress")
	options.SetDefault("Auth", true)
	options.SetDefault("MaxSize", 100*1024*1024)
	options.SetDefault("OpenshiftBuildCommit", "notrunninginopenshift")
	options.SetDefault("ValidTopics", "unit")
	options.SetDefault("Profile", false)
	options.SetDefault("Debug", false)
	options.SetDefault("DebugUserAgent", `unspecified`)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()
	kubenv := viper.New()
	kubenv.SetDefault("Openshift_Build_Commit", "notrunninginopenshift")
	kubenv.SetDefault("Hostname", "Hostname_Unavailable")
	kubenv.AutomaticEnv()

	return &IngressConfig{
		Hostname:             kubenv.GetString("Hostname"),
		MaxSize:              options.GetInt64("MaxSize"),
		StageBucket:          options.GetString("StageBucket"),
		Auth:                 options.GetBool("Auth"),
		KafkaBrokers:         options.GetStringSlice("KafkaBrokers"),
		KafkaGroupID:         options.GetString("KafkaGroupID"),
		KafkaTrackerTopic:    options.GetString("KafkaTrackerTopic"),
		ValidTopics:          strings.Split(options.GetString("ValidTopics"), ","),
		WebPort:              options.GetInt("WebPort"),
		MetricsPort:          options.GetInt("MetricsPort"),
		Profile:              options.GetBool("Profile"),
		Debug:                options.GetBool("Debug"),
		DebugUserAgent:       regexp.MustCompile(options.GetString("DebugUserAgent")),
		OpenshiftBuildCommit: kubenv.GetString("Openshift_Build_Commit"),
		Version:              "1.0.8",
		MinioEndpoint:        options.GetString("MinioEndpoint"),
		MinioAccessKey:       options.GetString("MinioAccessKey"),
		MinioSecretKey:       options.GetString("MinioSecretKey"),
		UseSSL:               options.GetBool("UseSSL"),
		UseClowder:           os.Getenv("CLOWDER_ENABLED") == "true",
	}
}
