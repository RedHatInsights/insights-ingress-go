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
	DefaultMaxSize       int64
	MaxSizeMap           map[string]string
	MaxUploadMem         int64
	Auth                 bool
    KafkaConfig          KafkaCfg
	WebPort              int
	MetricsPort          int
	Profile              bool
	OpenshiftBuildCommit string
	Version              string
	PayloadTrackerURL    string
	StorageConfig        StorageCfg
	LoggingConfig        LoggingCfg
	FeatureFlagsConfig   FeatureFlagCfg
	Debug                bool
	DebugUserAgent       *regexp.Regexp
}

type KafkaCfg struct {
	KafkaBrokers         []string
	KafkaGroupID         string
	KafkaTrackerTopic    string
	KafkaDeliveryReports bool
	KafkaAnnounceTopic   string
	ValidTopics          []string
	KafkaSSLConfig	     KafkaSSLCfg
}

type KafkaSSLCfg struct {
	KafkaCA			  string
	KafkaUsername	  string
	KafkaPassword	  string
	SASLMechanism	  string
	Protocol		  string
}

type FeatureFlagCfg struct {
	FFHostname string
	FFPort     int
	FFToken    string
	FFScheme   string
}

type StorageCfg struct {
	StageBucket      string
	StorageEndpoint  string
	StorageAccessKey string
	StorageSecretKey string
	UseSSL           bool
}

type LoggingCfg struct {
	LogGroup           string
	LogLevel           string
	AwsRegion          string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
}

// Get provides the IngressConfig
func Get() *IngressConfig {

	options := viper.New()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// global logging
	options.SetDefault("logLevel", "INFO")
	options.SetDefault("Hostname", hostname)

	// Kafka config
	options.SetDefault("KafkaGroupID", "ingress")
	options.SetDefault("KafkaDeliveryReports", true)
	options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	options.SetDefault("KafakAnnounceTopic", "platform.upload.announce")

	// Global defaults
	options.SetDefault("MaxUploadMem", 1024*1024*8)
	options.SetDefault("PayloadTrackerURL", "http://payload-tracker/v1/payloads/")
	options.SetDefault("Auth", true)
	options.SetDefault("DefaultMaxSize", 100*1024*1024)
	options.SetDefault("MaxSizeMap", `{}`)
	options.SetDefault("OpenshiftBuildCommit", "notrunninginopenshift")
	options.SetDefault("ValidTopics", "unit,announce")
	options.SetDefault("Profile", false)
	options.SetDefault("Debug", false)
	options.SetDefault("DebugUserAgent", `unspecified`)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()
	kubenv := viper.New()
	kubenv.SetDefault("Openshift_Build_Commit", "notrunninginopenshift")
	kubenv.AutomaticEnv()

	if clowder.IsClowderEnabled() {
		cfg := clowder.LoadedConfig

		sb := os.Getenv("INGRESS_STAGEBUCKET")
		bucket := clowder.ObjectBuckets[sb]
		broker := cfg.Kafka.Brokers[0]

		// Kafka
		options.SetDefault("KafkaBrokers", clowder.KafkaServers)
		options.SetDefault("KafkaTrackerTopic", clowder.KafkaTopics["platform.payload-status"].Name) 
		// Kafka SSL Config
		if broker.Authtype != nil {
			options.Set("KafkaUsername", *broker.Sasl.Username)
			options.Set("KafkaPassword", *broker.Sasl.Password)
			options.Set("SASLMechanism", "SCRAM-SHA-512")
			options.Set("Protocol", "sasl_ssl")
			caPath, err := cfg.KafkaCa(broker)
			if err != nil {
				panic("Kafka CA failed to write")
			}
			options.Set("KafkaCA", caPath)
		}
		// Ports
		options.SetDefault("WebPort", cfg.PublicPort)
		options.SetDefault("MetricsPort", cfg.MetricsPort)
		// Storage
		options.SetDefault("StageBucket", bucket.RequestedName)
		options.SetDefault("MinioEndpoint", fmt.Sprintf("%s:%d", cfg.ObjectStore.Hostname, cfg.ObjectStore.Port))
		options.SetDefault("MinioAccessKey", cfg.ObjectStore.Buckets[0].AccessKey)
		options.SetDefault("MinioSecretKey", cfg.ObjectStore.Buckets[0].SecretKey)
		options.SetDefault("UseSSL", cfg.ObjectStore.Tls)
		// Cloudwatch
		options.SetDefault("logGroup", cfg.Logging.Cloudwatch.LogGroup)
		options.SetDefault("AwsRegion", cfg.Logging.Cloudwatch.Region)
		options.SetDefault("AwsAccessKeyId", cfg.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("AwsSecretAccessKey", cfg.Logging.Cloudwatch.SecretAccessKey)
		// FeatureFlags
		options.SetDefault("FFHostname", cfg.FeatureFlags.Hostname)
		options.SetDefault("FFPort", cfg.FeatureFlags.Port)
		options.SetDefault("FFScheme", cfg.FeatureFlags.Scheme)
		options.SetDefault("FFToken", cfg.FeatureFlags.ClientAccessToken)
	} else {
		// Kafka
		defaultBrokers := os.Getenv("INGRESS_KAFKA_BROKERS")
		if len(defaultBrokers) == 0 {
			defaultBrokers = "kafka:29092"
		}
		options.SetDefault("KafkaBrokers", []string{defaultBrokers})
		options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
		// Ports
		options.SetDefault("WebPort", 3000)
		options.SetDefault("MetricsPort", 8080)
		// Storage
		options.SetDefault("StageBucket", "available")
		// Cloudwatch
		options.SetDefault("LogGroup", "platform-dev")
		options.SetDefault("AwsRegion", "us-east-1")
		options.SetDefault("UseSSL", false)
		options.SetDefault("AwsAccessKeyId", os.Getenv("CW_AWS_ACCESS_KEY_ID"))
		options.SetDefault("AwsSecretAccessKey", os.Getenv("CW_AWS_SECRET_ACCESS_KEY"))
	}

	IngressCfg := &IngressConfig{
		Hostname: options.GetString("Hostname"),
		DefaultMaxSize: options.GetInt64("DefaultMaxSize"),
		MaxSizeMap: options.GetStringMapString("MaxSizeMap"),
		MaxUploadMem: options.GetInt64("MaxUploadMem"),
		Auth: options.GetBool("Auth"),
		WebPort: options.GetInt("WebPort"),
		MetricsPort: options.GetInt("MetricsPort"),
		OpenshiftBuildCommit: kubenv.GetString("Openshift_Build_Commit"),
		Version: os.Getenv("1.0.8"),
		PayloadTrackerURL: options.GetString("PayloadTrackerURL"),
		Profile: options.GetBool("Profile"),
		Debug: options.GetBool("Debug"),
		DebugUserAgent: regexp.MustCompile(options.GetString("DebugUserAgent")),
		KafkaConfig: KafkaCfg{
			KafkaBrokers:         options.GetStringSlice("KafkaBrokers"),
			KafkaGroupID:         options.GetString("KafkaGroupID"),
			KafkaTrackerTopic:    options.GetString("KafkaTrackerTopic"),
			KafkaDeliveryReports: options.GetBool("KafkaDeliveryReports"),
			KafkaAnnounceTopic:   options.GetString("KafakAnnounceTopic"),
			ValidTopics: 		  strings.Split(options.GetString("ValidTopics"), ","),
		},
		StorageConfig: StorageCfg{
			StageBucket: options.GetString("StageBucket"),
			StorageEndpoint: options.GetString("MinioEndpoint"),
			StorageAccessKey: options.GetString("MinioAccessKey"),
			StorageSecretKey: options.GetString("MinioSecretKey"),
			UseSSL: options.GetBool("UseSSL"),
		},
		LoggingConfig: LoggingCfg{
			LogGroup: options.GetString("logGroup"),
			LogLevel: options.GetString("logLevel"),
			AwsRegion: options.GetString("AwsRegion"),
			AwsAccessKeyId: options.GetString("AwsAccessKeyId"),
			AwsSecretAccessKey: options.GetString("AwsSecretAccessKey"),
		},
	}

	if options.IsSet("KafkaUsername") {
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaUsername = options.GetString("KafkaUsername")
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaPassword = options.GetString("KafkaPassword")
		IngressCfg.KafkaConfig.KafkaSSLConfig.SASLMechanism = options.GetString("SASLMechanism")
		IngressCfg.KafkaConfig.KafkaSSLConfig.Protocol = options.GetString("Protocol")
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaCA = options.GetString("KafkaCA")
	}
	
	if options.IsSet("FFHostname") {
		IngressCfg.FeatureFlagsConfig.FFHostname = options.GetString("FFHostname")
		IngressCfg.FeatureFlagsConfig.FFToken = options.GetString("FFToken")
		IngressCfg.FeatureFlagsConfig.FFPort = options.GetInt("FFPort")
		IngressCfg.FeatureFlagsConfig.FFScheme = options.GetString("FFScheme")
	}

	return IngressCfg
}
