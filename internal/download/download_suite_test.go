package download_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestDownload(t *testing.T) {
	cfg := config.Get()
	RegisterFailHandler(Fail)
	l.InitLogger(cfg)
	RunSpecs(t, "Test Download Suite")
}
