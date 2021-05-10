package upload_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	l "github.com/redhatinsights/insights-ingress-go/pkg/logger"
)

func TestUpload(t *testing.T) {
	RegisterFailHandler(Fail)
	l.InitLogger()
	RunSpecs(t, "Upload Suite")
}
