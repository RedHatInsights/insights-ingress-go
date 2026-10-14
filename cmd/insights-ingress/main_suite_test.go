package main

import (
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

func TestIngress(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	l.InitLogger(config.Get())
	ginkgo.RunSpecs(t, "Ingress Suite")
}
