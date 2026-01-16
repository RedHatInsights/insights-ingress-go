package kafka_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/insights-ingress-go/internal/validators"
	. "github.com/redhatinsights/insights-ingress-go/internal/validators/kafka"
)

var _ = Describe("Kafka", func() {
	Describe("Validating a service", func() {
		Context("that has a valid topic", func() {
			It("should not error", func() {
				kv := buildValidator([]string{"test1", "test2", "unit"})
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unit",
					Category: "test",
				})
				Expect(err).To(BeNil())
			})
		})

		Context("that has a valid topic and exists in the mapping", func() {
			It("should not error", func() {
				kv := buildValidator([]string{"test1", "test2", "unit", "unit2"})
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unit2",
					Category: "test",
				})
				Expect(err).To(BeNil())
			})
		})

		Context("that does not have a valid topic", func() {
			It("should error", func() {
				kv := buildValidator([]string{"test1", "test2", "unit"})
				err := kv.ValidateService(&validators.ServiceDescriptor{
					Service:  "unknown",
					Category: "test",
				})
				Expect(err.Error()).To(Equal("Upload type is not supported: unknown"))
			})
		})
	})
})

func buildValidator(validUploadTypes []string) *Validator {
	kafkaCfg := Config{Brokers: []string{"broker1"}}
	return New(&kafkaCfg, validUploadTypes...)
}
