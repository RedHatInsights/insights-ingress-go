package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/health"
)

var _ = Describe("readiness handler", func() {
	It("returns 200 and each dependency status when all checks pass", func() {
		handler := health.Handler([]health.Dependency{
			{Name: "kafka", Checker: health.CheckFunc(func(context.Context) error { return nil })},
			{Name: "storage", Backend: "s3", Checker: health.CheckFunc(func(context.Context) error { return nil })},
		}, time.Second)

		recorder := httptest.NewRecorder()
		handler(recorder, httptest.NewRequest(http.MethodGet, "/status/", nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Header().Get("Content-Type")).To(Equal("application/json"))
		Expect(recorder.Body.String()).To(ContainSubstring(`"status":"ok"`))
		Expect(recorder.Body.String()).To(ContainSubstring(`"kafka":{"status":"ok"}`))
		Expect(recorder.Body.String()).To(ContainSubstring(`"storage":{"status":"ok","backend":"s3"}`))
	})

	It("returns 503 when a critical dependency fails", func() {
		handler := health.Handler([]health.Dependency{
			{Name: "kafka", Checker: health.CheckFunc(func(context.Context) error { return errors.New("broker credentials should not be exposed") })},
			{Name: "storage", Checker: health.CheckFunc(func(context.Context) error { return nil })},
		}, time.Second)

		recorder := httptest.NewRecorder()
		handler(recorder, httptest.NewRequest(http.MethodGet, "/status/", nil))

		Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(recorder.Body.String()).To(ContainSubstring(`"status":"error"`))
		Expect(recorder.Body.String()).To(ContainSubstring(`"kafka":{"status":"error","error":"dependency check failed"}`))
		Expect(recorder.Body.String()).ToNot(ContainSubstring("broker credentials"))
	})

	It("returns 503 when a dependency exceeds the readiness deadline", func() {
		handler := health.Handler([]health.Dependency{
			{Name: "kafka", Checker: health.CheckFunc(func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			})},
		}, 10*time.Millisecond)

		recorder := httptest.NewRecorder()
		started := time.Now()
		handler(recorder, httptest.NewRequest(http.MethodGet, "/status/", nil))

		Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(time.Since(started)).To(BeNumerically("<", time.Second))
		Expect(recorder.Body.String()).To(ContainSubstring(`"error":"dependency check timed out"`))
	})
})
