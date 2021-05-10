package version_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	l "github.com/redhatinsights/insights-ingress-go/pkg/logger"
)

func TestInventory(t *testing.T) {
	RegisterFailHandler(Fail)
	l.InitLogger()
	RunSpecs(t, "Version Suite")
}
