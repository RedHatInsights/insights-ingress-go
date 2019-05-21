package s3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "cloud.redhat.com/ingress/stage/s3"
)

var _ = Describe("S3", func() {
	Describe("A valid s3 url", func() {
		It("should return a valid S3Spec", func() {
			spec, err := FromURL("https://test.s3.amazonaws.com/my-key")
			Expect(err).To(BeNil())
			Expect(spec.Bucket).To(Equal("test"))
			Expect(spec.Key).To(Equal("my-key"))
		})
	})

	Describe("An invalid s3 url", func() {
		It("Should return an error", func() {
			spec, err := FromURL("https://localhost/?q=foo")
			Expect(err).To(Not(BeNil()))
			Expect(spec).To(BeNil())
		})
	})
})
