package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	rhiconfig "github.com/redhatinsights/app-common-go/pkg/api/v1"

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
	HTTPClientTimeout    time.Duration
	Profile              bool
	OpenshiftBuildCommit string
	Version              string
	PayloadTrackerURL    string
	TlsCAPath            string
	StorageConfig        StorageCfg
	LoggingConfig        LoggingCfg
	DenyListedOrgIDs     []string
	Debug                bool
	DebugUserAgent       *regexp.Regexp
	ServiceBaseURL       string
	StagerImplementation string
}

type KafkaCfg struct {
	KafkaBrokers          []string
	KafkaGroupID          string
	KafkaTrackerTopic     string
	KafkaDeliveryReports  bool
	KafkaAnnounceTopic    string
	ValidUploadTypes      []string
	KafkaSecurityProtocol string
	KafkaSSLConfig        KafkaSSLCfg
}

type KafkaSSLCfg struct {
	KafkaCA       string
	KafkaUsername string
	KafkaPassword string
	SASLMechanism string
}

type StorageCfg struct {
	StageBucket           string
	StorageEndpoint       string
	StorageAccessKey      string
	StorageSecretKey      string
	UseSSL                bool
	StorageRegion         string
	StorageFileSystemPath string
}

type LoggingCfg struct {
	LogGroup           string
	LogLevel           string
	AwsRegion          string
	AwsAccessKeyId     string
	AwsSecretAccessKey string
}

func GetTopic(topic string) string {
	if clowder.IsClowderEnabled() {
		return rhiconfig.KafkaTopics[topic].Name
	}
	return topic
}

func GetStagerImplementation(stagerImplementation string, stageFileSystemPath string) string {
	if !clowder.IsClowderEnabled() {
		lowerStagerImplementation := strings.ToLower(stagerImplementation)
		if lowerStagerImplementation == "filebased" {
			if len(stageFileSystemPath) != 0 {
				err := os.MkdirAll(stageFileSystemPath, os.ModePerm)
				if err == nil {
					return "filebased"
				}
			}
		}
	}
	return "s3"
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
	options.SetDefault("KafkaBrokers", []string{"kafka:29092"})
	options.SetDefault("KafkaGroupID", "ingress")
	options.SetDefault("KafkaDeliveryReports", true)
	options.SetDefault("KafkaTrackerTopic", "platform.payload-status")
	options.SetDefault("KafkaAnnounceTopic", "platform.upload.announce")
	options.SetDefault("KafkaSecurityProtocol", "PLAINTEXT")

	// Storage (MinIO/S3-compatible)
	options.SetDefault("StageBucket", "available")
	options.SetDefault("MinioEndpoint", "")
	options.SetDefault("MinioAccessKey", "")
	options.SetDefault("MinioSecretKey", "")
	options.SetDefault("StorageRegion", "")
	options.SetDefault("UseSSL", false)

	// Cloudwatch
	options.SetDefault("LogGroup", "")
	options.SetDefault("AwsRegion", "")
	options.SetDefault("AwsAccessKeyId", "")
	options.SetDefault("AwsSecretAccessKey", "")

	// Global defaults
	options.SetDefault("WebPort", 3000)
	options.SetDefault("MetricsPort", 8080)
	options.SetDefault("MaxUploadMem", 1024*1024*8)
	options.SetDefault("PayloadTrackerURL", "http://payload-tracker/v1/payloads/")
	options.SetDefault("TlsCAPath", "")
	options.SetDefault("HTTPClientTimeout", 10)
	options.SetDefault("Auth", true)
	options.SetDefault("DefaultMaxSize", 100*1024*1024)
	options.SetDefault("MaxSizeMap", `{}`)
	options.SetDefault("OpenshiftBuildCommit", "notrunninginopenshift")
	options.SetDefault("Valid_Upload_Types", "unit,announce")
	options.SetDefault("Profile", false)
	options.SetDefault("Deny_Listed_OrgIDs", []string{})
	options.SetDefault("Debug", false)
	options.SetDefault("DebugUserAgent", `unspecified`)
	options.SetDefault("ServiceBaseURL", "http://localhost:3000")
	options.SetDefault("StagerImplementation", "s3")
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
		options.SetDefault("KafkaAnnounceTopic", clowder.KafkaTopics["platform.upload.announce"].Name)

		if broker.SecurityProtocol != nil && *broker.SecurityProtocol != "" {
			options.Set("KafkaSecurityProtocol", *broker.SecurityProtocol)
		} else if broker.Sasl != nil && broker.Sasl.SecurityProtocol != nil && *broker.Sasl.SecurityProtocol != "" {
			options.Set("KafkaSecurityProtocol", *broker.Sasl.SecurityProtocol)
		}

		// Kafka SSL Config
		if broker.Authtype != nil {
			options.Set("KafkaUsername", *broker.Sasl.Username)
			options.Set("KafkaPassword", *broker.Sasl.Password)
			options.Set("SASLMechanism", *broker.Sasl.SaslMechanism)
		}
		if broker.Cacert != nil {
			caPath, err := cfg.KafkaCa(broker)
			if err != nil {
				panic("Kafka CA failed to write")
			}
			options.Set("KafkaCA", caPath)
		}

		// TLS
		options.SetDefault("TlsCAPath", cfg.TlsCAPath)
		// Ports
		options.SetDefault("WebPort", cfg.PublicPort)
		options.SetDefault("MetricsPort", cfg.MetricsPort)
		// Storage
		options.SetDefault("StageBucket", bucket.Name)
		options.SetDefault("MinioEndpoint", fmt.Sprintf("%s:%d", cfg.ObjectStore.Hostname, cfg.ObjectStore.Port))
		options.SetDefault("MinioAccessKey", cfg.ObjectStore.Buckets[0].AccessKey)
		options.SetDefault("MinioSecretKey", cfg.ObjectStore.Buckets[0].SecretKey)
		options.SetDefault("UseSSL", cfg.ObjectStore.Tls)
		if cfg.ObjectStore.Buckets[0].Region != nil {
			options.SetDefault("StorageRegion", cfg.ObjectStore.Buckets[0].Region)
		} else {
			options.SetDefault("StorageRegion", "")
		}
		// Cloudwatch
		options.SetDefault("logGroup", cfg.Logging.Cloudwatch.LogGroup)
		options.SetDefault("AwsRegion", cfg.Logging.Cloudwatch.Region)
		options.SetDefault("AwsAccessKeyId", cfg.Logging.Cloudwatch.AccessKeyId)
		options.SetDefault("AwsSecretAccessKey", cfg.Logging.Cloudwatch.SecretAccessKey)
	}

	IngressCfg := &IngressConfig{
		Hostname:             options.GetString("Hostname"),
		DefaultMaxSize:       options.GetInt64("DefaultMaxSize"),
		MaxSizeMap:           options.GetStringMapString("MaxSizeMap"),
		MaxUploadMem:         options.GetInt64("MaxUploadMem"),
		Auth:                 options.GetBool("Auth"),
		WebPort:              options.GetInt("WebPort"),
		MetricsPort:          options.GetInt("MetricsPort"),
		HTTPClientTimeout:    time.Duration(options.GetInt("HTTPClientTimeout")),
		OpenshiftBuildCommit: kubenv.GetString("Openshift_Build_Commit"),
		Version:              os.Getenv("1.0.8"),
		PayloadTrackerURL:    options.GetString("PayloadTrackerURL"),
		TlsCAPath:            options.GetString("TlsCAPath"),
		Profile:              options.GetBool("Profile"),
		DenyListedOrgIDs:     options.GetStringSlice("Deny_Listed_OrgIDs"),
		Debug:                options.GetBool("Debug"),
		DebugUserAgent:       regexp.MustCompile(options.GetString("DebugUserAgent")),
		KafkaConfig: KafkaCfg{
			KafkaBrokers:          options.GetStringSlice("KafkaBrokers"),
			KafkaGroupID:          options.GetString("KafkaGroupID"),
			KafkaTrackerTopic:     options.GetString("KafkaTrackerTopic"),
			KafkaDeliveryReports:  options.GetBool("KafkaDeliveryReports"),
			KafkaAnnounceTopic:    options.GetString("KafkaAnnounceTopic"),
			ValidUploadTypes:      strings.Split(options.GetString("Valid_Upload_Types"), ","),
			KafkaSecurityProtocol: options.GetString("KafkaSecurityProtocol"),
		},
		StorageConfig: StorageCfg{
			StageBucket:           options.GetString("StageBucket"),
			StorageEndpoint:       options.GetString("MinioEndpoint"),
			StorageAccessKey:      options.GetString("MinioAccessKey"),
			StorageSecretKey:      options.GetString("MinioSecretKey"),
			UseSSL:                options.GetBool("UseSSL"),
			StorageRegion:         options.GetString("StorageRegion"),
			StorageFileSystemPath: options.GetString("StorageFileSystemPath"),
		},
		LoggingConfig: LoggingCfg{
			LogGroup:           options.GetString("logGroup"),
			LogLevel:           options.GetString("logLevel"),
			AwsRegion:          options.GetString("AwsRegion"),
			AwsAccessKeyId:     options.GetString("AwsAccessKeyId"),
			AwsSecretAccessKey: options.GetString("AwsSecretAccessKey"),
		},
		ServiceBaseURL:       options.GetString("ServiceBaseURL"),
		StagerImplementation: options.GetString("StagerImplementation"),
	}

	if options.IsSet("KafkaUsername") {
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaUsername = options.GetString("KafkaUsername")
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaPassword = options.GetString("KafkaPassword")
		IngressCfg.KafkaConfig.KafkaSSLConfig.SASLMechanism = options.GetString("SASLMechanism")
	}

	if options.IsSet("KafkaCA") {
		IngressCfg.KafkaConfig.KafkaSSLConfig.KafkaCA = options.GetString("KafkaCA")
	}

	return IngressCfg
}
