package kafka_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	l "github.com/redhatinsights/insights-ingress-go/logger"
)

func TestKafka(t *testing.T) {
	RegisterFailHandler(Fail)
	l.InitLogger()
	RunSpecs(t, "Kafka Suite")
}
