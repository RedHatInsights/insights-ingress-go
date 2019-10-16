package logger

import (
	"flag"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	logrus_cloudwatchlogs "github.com/kdar/logrus-cloudwatchlogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Log is an instance of the global logrus.Logger
var Log *logrus.Logger
var logLevel logrus.Level

// InitLogger initializes the Entitlements API logger
func InitLogger() *logrus.Logger {

	viper.SetDefault("INGRESS_LOG_LEVEL", "INFO")
	viper.SetDefault("INGRESS_LOG_GROUP", "platform-dev")
	viper.SetDefault("INGRESS_LOG_STREAM", "platform")
	viper.SetDefault("INGRESS_AWS_REGION", "us-east-1")
	key := viper.GetString("INGRESS_CW_AWS_ACCESS_KEY_ID")
	secret := viper.GetString("INGRESS_CW_AWS_SECRET_ACCESS_KEY")
	region := viper.GetString("INGRESS_AWS_REGION")
	group := viper.GetString("INGRESS_LOG_GROUP")
	stream := viper.GetString("INGRESS_LOG_STREAM")
	viper.AutomaticEnv()
	switch viper.GetString("INGRESS_LOG_LEVEL") {
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

	Log = &logrus.Logger{
		Out:   os.Stdout,
		Level: logLevel,
	}

	formatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyFunc:  "caller",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
		},
	}

	Log.SetFormatter(formatter)

	cred := credentials.NewStaticCredentials(key, secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(cred)

	if key != "" {
		hook, err := logrus_cloudwatchlogs.NewHook(group, stream, cfg)
		if err != nil {
			Log.Info(err)
		}
		Log.AddHook(hook)
	}

	return Log
}
