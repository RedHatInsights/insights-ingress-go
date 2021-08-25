package track_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestTrack(t *testing.T) {
	RegisterFailHandler(Fail)
	l.InitLogger()
	RunSpecs(t, "Track Suite")
}
