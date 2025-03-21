package filebased_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestFileBased(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.Get()
	cfg.StagerImplementation = "filebased"
	cfg.StorageConfig.StorageFileSystemPath = tempDir
	RegisterFailHandler(Fail)
	l.InitLogger(cfg)
	RunSpecs(t, "Test File Based Suite")
}
