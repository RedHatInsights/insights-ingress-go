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
	DefaultMaxSize		 int64
	MaxSizeMap			 map[string]string
	StageBucket          string
	Auth                 bool
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaTrackerTopic    string
	KafkaCA				 string
	KafkaUsername		 string
	KafkaPassword		 string
	DeliveryReports	   	 bool
	SASLMechanism		 string
	Protocol             string
	ValidTopics          []string
	WebPort              int
	MetricsPort          int
	Profile              bool
	OpenshiftBuildCommit string
	Version              string
	PayloadTrackerURL  string
	MinioEndpoint        string
	MinioAccessKey       string
	MinioSecretKey       string
	LogGroup             string
	LogLevel             string
	AwsRegion            string
	AwsAccessKeyId       string
	AwsSecretAccessKey   string
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
		bucket := clowder.ObjectBuckets[sb]

		options.SetDefault("WebPort", cfg.PublicPort)
		options.SetDefault("MetricsPort", cfg.MetricsPort)
		options.SetDefault("KafkaBrokers", clowder.KafkaServers)
		options.SetDefault("MinioEndpoint", fmt.Sprintf("%s:%d", cfg.ObjectStore.Hostname, cfg.ObjectStore.Port))
		options.SetDefault("MinioAccessKey", *cfg.ObjectStore.Buckets[0].AccessKey)
		options.SetDefault("MinioSecretKey", *cfg.ObjectStore.Buckets[0].SecretKey)
		options.SetDefault("UseSSL", cfg.ObjectStore.Tls)
		options.SetDefault("StageBucket", bucket.RequestedName)
		options.SetDefault("LogGroup", cfg.Logging.Cloudwatch.LogGroup)
		options.SetDefault("AwsRegion", cfg.Logging.Cloudwatch.Region)
		options.SetDefault("AwsAccessKeyId", cfg.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("AwsSecretAccessKey", cfg.Logging.Cloudwatch.SecretAccessKey)
		options.SetDefault("KafkaTrackerTopic", clowder.KafkaTopics["platform.payload-status"].Name)
	} else {
		options.SetDefault("WebPort", 3000)
		options.SetDefault("MetricsPort", 8080)
		options.SetDefault("KafkaBrokers", []string{"kafka:29092"})
		options.SetDefault("StageBucket", "available")
		options.SetDefault("LogGroup", "platform-dev")
		options.SetDefault("AwsRegion", "us-east-1")
		options.SetDefault("UseSSL", false)
		options.SetDefault("AwsAccessKeyId", os.Getenv("CW_AWS_ACCESS_KEY_ID"))
		options.SetDefault("AwsSecretAccessKey", os.Getenv("CW_AWS_SECRET_ACCESS_KEY"))
		options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	}

	options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	options.SetDefault("KafkaGroupID", "ingress")
	options.SetDefault("DeliveryReports", false)
	options.SetDefault("LogLevel", "INFO")
	options.SetDefault("PayloadTrackerURL", "http://payload-tracker/v1/payloads/")
	options.SetDefault("Auth", true)
	options.SetDefault("DefaultMaxSize", 100*1024*1024)
	options.SetDefault("MaxSizeMap", `{}`)
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

	ingressCfg := &IngressConfig{
		Hostname:             kubenv.GetString("Hostname"),
		DefaultMaxSize:       options.GetInt64("DefaultMaxSize"),
		MaxSizeMap:			  options.GetStringMapString("MaxSizeMap"),
		StageBucket:          options.GetString("StageBucket"),
		Auth:                 options.GetBool("Auth"),
		KafkaBrokers:         options.GetStringSlice("KafkaBrokers"),
		KafkaGroupID:         options.GetString("KafkaGroupID"),
		KafkaTrackerTopic:    options.GetString("KafkaTrackerTopic"),
		DeliveryReports:       options.GetBool("DeliveryReports"),
		ValidTopics:          strings.Split(options.GetString("ValidTopics"), ","),
		WebPort:              options.GetInt("WebPort"),
		MetricsPort:          options.GetInt("MetricsPort"),
		PayloadTrackerURL:   options.GetString("PayloadTrackerURL"),
		Profile:              options.GetBool("Profile"),
		Debug:                options.GetBool("Debug"),
		DebugUserAgent:       regexp.MustCompile(options.GetString("DebugUserAgent")),
		OpenshiftBuildCommit: kubenv.GetString("Openshift_Build_Commit"),
		Version:              "1.0.8",
		MinioEndpoint:        options.GetString("MinioEndpoint"),
		MinioAccessKey:       options.GetString("MinioAccessKey"),
		MinioSecretKey:       options.GetString("MinioSecretKey"),
		LogGroup:             options.GetString("LogGroup"),
		LogLevel:             options.GetString("LogLevel"),
		AwsRegion:            options.GetString("AwsRegion"),
		AwsAccessKeyId:       options.GetString("AwsAccessKeyId"),
		AwsSecretAccessKey:   options.GetString("AwsSecretAccessKey"),
		UseSSL:               options.GetBool("UseSSL"),
		UseClowder:           os.Getenv("CLOWDER_ENABLED") == "true",
	}
	
	if os.Getenv("CLOWDER_ENABLED") == "true" {
		cfg := clowder.LoadedConfig
		broker := cfg.Kafka.Brokers[0]

		if broker.Authtype != nil {
			ingressCfg.KafkaUsername = *broker.Sasl.Username
			ingressCfg.KafkaPassword = *broker.Sasl.Password
			ingressCfg.SASLMechanism = "SCRAM-SHA-512"
			ingressCfg.Protocol = "sasl_ssl"
			caPath, err := cfg.KafkaCa(broker)

			if err != nil {
				panic("Kafka CA failed to write")
			}

			ingressCfg.KafkaCA = caPath
		}
	}

	return ingressCfg

}
