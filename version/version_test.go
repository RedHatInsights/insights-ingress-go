package version_test

import (
	"github.com/redhatinsights/insights-ingress-go/version"

	"errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func GetServer() (*httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		errors.New("version endpoint failed")
	}

	rr := httptest.NewRecorder()

	return rr, req

}

var _ = Describe("Version", func() {

	Describe("Reading from the version file", func() {
		It("should return a string", func() {
			s := version.ReadVersion()
			Expect(s).NotTo(BeNil())
			Expect(s).To(Equal("0.0.0"))
		})
	})

	Describe("GET from the version endpoint", func() {
		It("should return a json doc containg version", func() {
			rr, req := GetServer()
			handler := http.HandlerFunc(version.GetVersion)
			handler.ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal(`{"version":"0.0.0","commit":"notrunninginopenshift"}`))
		})
	})
})
