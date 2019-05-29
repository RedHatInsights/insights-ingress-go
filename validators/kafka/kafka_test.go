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
	})
})
