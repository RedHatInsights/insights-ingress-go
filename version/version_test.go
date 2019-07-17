package version_test

import (
	"github.com/redhatinsights/insights-ingress-go/version"

	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GetServer() (*httptest.ResponseRecorder, *http.Request) {
	req, _ := http.NewRequest("GET", "/api/ingress/v1/version", nil)
	rr := httptest.NewRecorder()

	return rr, req

}

var _ = Describe("Version", func() {

	Describe("GET from the version endpoint", func() {
		It("should return a json doc containg version", func() {
			rr, req := GetServer()
			handler := http.HandlerFunc(version.GetVersion)
			handler.ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal(`{"version":"1.0.1","commit":"notrunninginopenshift"}`))
		})
	})
})
