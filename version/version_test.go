package version_test

import (
	"github.com/redhatinsights/insights-ingress-go/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {

	Describe("Reading from the version file", func() {
		It("Should return a string", func() {
			s := version.ReadVersion()
			Expect(s).NotTo(BeNil())
			Expect(s).To(Equal("1.0.0"))
		})
	})
})
