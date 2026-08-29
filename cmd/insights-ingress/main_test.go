package main

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("liveness endpoint", func() {
	It("returns success without invoking dependencies", func() {
		recorder := httptest.NewRecorder()
		lubDub(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(Equal("lubdub"))
	})
})
