package kafka_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/insights-ingress-go/validators"
	. "github.com/redhatinsights/insights-ingress-go/validators/kafka"
)

var _ = Describe("Kafka", func() {
	var (
		kv *Validator
	)

	Describe("Validating a service", func() {
		Context("that has a valid topic", func() {
			It("should not error", func() {
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unit",
					Category: "test",
				})
				Expect(err).To(BeNil())
			})
		})

		Context("that has a valid topic and exists in the mapping", func() {
			It("should not error", func() {
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unit2",
					Category: "test",
				})
				Expect(err).To(BeNil())
			})
		})

		Context("that does not have a valid topic", func() {
			It("should error", func() {
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unknown",
					Category: "test",
				})
				Expect(err.Error()).To(Equal("Validation topic is invalid"))
			})
		})
	})
})
