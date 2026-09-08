package s3compat_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestS3Compat(t *testing.T) {
	RegisterFailHandler(Fail)
	l.InitLogger(config.Get())
	RunSpecs(t, "S3 Compat Suite")
}
