package logger

import (
	"flag"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyFunc:  "caller",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	}

	Log = &logrus.Logger{
		Out:       os.Stdout,
		Level:     logLevel,
		Formatter: formatter,
	}

	cred := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(cred)

	if key != "" {
		hook, err := lc.NewHook(group, stream, cfg)
		if err != nil {
			Log.Info(err)
		} else {
			Log.AddHook(hook)
		}
	}

	return Log
}
