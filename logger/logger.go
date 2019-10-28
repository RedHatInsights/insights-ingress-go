package logger

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	lc "github.com/kdar/logrus-cloudwatchlogs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Log is an instance of the global logrus.Logger
var Log *logrus.Logger
var logLevel logrus.Level

// NewCloudwatchFormatter creates a new log formatter
func NewCloudwatchFormatter() *CustomCloudwatch {
	f := &CustomCloudwatch{}

	var err error
	if f.Hostname == "" {
		if f.Hostname, err = os.Hostname(); err != nil {
			f.Hostname = "unknown"
		}
	}

	return f
}

//Format is the log formatter for the entry
func (f *CustomCloudwatch) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	now := time.Now()

	hostname, err := os.Hostname()
	if err == nil {
		f.Hostname = hostname
	}

	data := map[string]interface{}{
		"@timestamp":  now.Format("2006-01-02T15:04:05.999Z"),
		"@version":    1,
		"message":     entry.Message,
		"levelname":   entry.Level.String(),
		"source_host": f.Hostname,
		"app":         "ingress",
		"caller":      entry.Caller.Func.Name(),
	}

	j, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	b.Write(j)

	return b.Bytes(), nil
}

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

	formatter := NewCloudwatchFormatter()

	Log = &logrus.Logger{
		Out:          os.Stdout,
		Level:        logLevel,
		Formatter:    formatter,
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
