package kafka_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/redhatinsights/insights-ingress-go/validators"
	. "github.com/redhatinsights/insights-ingress-go/validators/kafka"
)

var _ = Describe("Kafka", func() {
	var (
		kv *Validator
	)

	var wait = func(ch chan *validators.Response) bool {
		select {
		case <-ch:
			return true
		case <-time.After(100 * time.Millisecond):
			return false
		}
	}

	BeforeEach(func() {
		kv = &Validator{
			ValidChan:   make(chan *validators.Response),
			InvalidChan: make(chan *validators.Response),
		}
	})

	Describe("Routing a response", func() {
		Context("with validation field set to 'success'", func() {
			It("should forward to the Valid channel", func() {
				go kv.RouteResponse(&validators.Response{
					Validation: "success",
					Timestamp:  time.Now(),
				})
				Expect(wait(kv.ValidChan)).To(BeTrue())
			})
		})

		Context("with validation field set to 'failure'", func() {
			It("should forward to the Invalid channel", func() {
				go kv.RouteResponse(&validators.Response{
					Validation: "failure",
				})
				Expect(wait(kv.InvalidChan)).To(BeTrue())
			})
		})

		Context("with validation field set to 'invalid'", func() {
			It("should not forward to any channel", func() {
				go kv.RouteResponse(&validators.Response{
					Validation: "invalid",
				})
				Expect(wait(kv.InvalidChan)).To(BeFalse())
				Expect(wait(kv.ValidChan)).To(BeFalse())
			})
		})
		Context("with validation set to 'handoff'", func() {
			It("should not forward to any channel", func() {
				go kv.RouteResponse(&validators.Response{
					Validation: "handoff",
				})
				Expect(wait(kv.InvalidChan)).To(BeFalse())
				Expect(wait(kv.ValidChan)).To(BeFalse())
			})
		})
	})

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

	Describe("Updating the extras field in a response", func() {
		Context("with an empty ID and SatelliteManaged field", func() {
			It("should copy the zero values to the Extras field", func() {
				vr := &validators.Response{}
				UpdateExtras(vr)
				Expect(vr.Extras["id"]).To(Equal(""))
				Expect(vr.Extras["satellite_managed"]).To(BeNil())
			})
		})

		Context("with a value for ID and SatelliteManaged", func() {
			It("should copy the values to the Extras field", func() {
				bp := true
				vr := &validators.Response{
					ID:               "testing",
					SatelliteManaged: &bp,
				}
				UpdateExtras(vr)
				Expect(vr.Extras["id"]).To(Equal("testing"))
				sm := vr.Extras["satellite_managed"].(*bool)
				Expect(*sm).To(BeTrue())
			})
		})

		Context("with extras already filled out", func() {
			It("should not overwrite the contents of extra", func() {
				vr := &validators.Response{
					Extras: map[string]interface{}{
						"id":                "testing",
						"satellite_managed": nil,
					},
				}
				UpdateExtras(vr)
				Expect(vr.Extras["id"]).To(Equal("testing"))
				Expect(vr.Extras["satellite_managed"]).To(BeNil())
			})
		})

		Context("with extras partially filled", func() {
			It("should fill in the remainder", func() {
				b := true
				vr := &validators.Response{
					SatelliteManaged: &b,
					Extras: map[string]interface{}{
						"id": "testing",
					},
				}
				UpdateExtras(vr)
				Expect(vr.Extras["id"]).To(Equal("testing"))

				sm := vr.Extras["satellite_managed"].(*bool)
				Expect(*sm).To(BeTrue())
			})
		})
	})
})
