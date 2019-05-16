package config

import (
	"github.com/spf13/viper"
)

// IngressConfig represents the runtime configuration
type IngressConfig struct {
	MaxSize            int
	StageBucket        string
	RejectBucket       string
	Auth               bool
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
}

// Get returns an initialized IngressConfig
func Get() *IngressConfig {
	options := viper.New()
	options.SetDefault("MaxSize", 10*1024*1024)
	options.SetDefault("StageBucket", "available")
	options.SetDefault("RejectBucket", "rejected")
	options.SetDefault("Auth", true)
	options.SetEnvPrefix("INGRESS")
	options.AutomaticEnv()

	return &IngressConfig{
		MaxSize:      options.GetInt("MaxSize"),
		StageBucket:  options.GetString("StageBucket"),
		RejectBucket: options.GetString("RejectBucket"),
	}
}
