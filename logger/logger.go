package logger

import (
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	logrustash "github.com/bshuster-repo/logrus-logstash-hook"
	lc "github.com/kdar/logrus-cloudwatchlogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Log is an instance of the global logrus.Logger
var Log *logrus.Logger
var logLevel logrus.Level

// InitLogger initializes the Entitlements API logger
func InitLogger() *logrus.Logger {

	logconfig := viper.New()
	logconfig.SetDefault("LOG_LEVEL", "INFO")
	logconfig.SetDefault("LOG_GROUP", "platform-dev")
	logconfig.SetDefault("LOG_STREAM", "platform")
	logconfig.SetDefault("AWS_REGION", "us-east-1")
	logconfig.SetEnvPrefix("INGRESS")
	logconfig.AutomaticEnv()
	key := logconfig.GetString("CW_AWS_ACCESS_KEY_ID")
	secret := logconfig.GetString("CW_AWS_SECRET_ACCESS_KEY")
	region := logconfig.GetString("AWS_REGION")
	group := logconfig.GetString("LOG_GROUP")
	stream := logconfig.GetString("LOG_STREAM")

	switch logconfig.GetString("LOG_LEVEL") {
	case "DEBUG":
		logLevel = logrus.DebugLevel
	case "ERROR":
		logLevel = logrus.ErrorLevel
	default:
		logLevel = logrus.InfoLevel
	}
	if flag.Lookup("test.v") != nil {
		logLevel = logrus.FatalLevel
	}

	Log = logrus.New()

	Log = &logrus.Logger{
		Out:          os.Stdout,
		Level:        logLevel,
		Formatter:    &logrustash.LogstashFormatter{},
		Hooks:        make(logrus.LevelHooks),
		ReportCaller: true,
	}

	cred := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(cred)

	if key != "" {
		hook, err := lc.NewBatchingHook(group, stream, cfg, 200*time.Millisecond)
		if err != nil {
			Log.Info(err)
		} else {
			Log.AddHook(hook)
		}
	}

	return Log
}
