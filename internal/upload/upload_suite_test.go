package upload_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestUpload(t *testing.T) {
	cfg := config.Get()
	RegisterFailHandler(Fail)
	l.InitLogger(cfg)
	RunSpecs(t, "Upload Suite")
}
