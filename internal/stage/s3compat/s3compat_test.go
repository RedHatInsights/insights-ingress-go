package s3compat_test

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/minio/minio-go/v6"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/stage/s3compat"
)

type blockingRoundTripper struct {
	requestStarted  chan struct{}
	requestCanceled chan struct{}
	startedOnce     sync.Once
	canceledOnce    sync.Once
}

func (t *blockingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	t.startedOnce.Do(func() { close(t.requestStarted) })
	<-request.Context().Done()
	t.canceledOnce.Do(func() { close(t.requestCanceled) })
	return nil, request.Context().Err()
}

var _ = Describe("S3 readiness check", func() {
	It("cancels an in-flight bucket check with the readiness context", func() {
		transport := &blockingRoundTripper{
			requestStarted:  make(chan struct{}),
			requestCanceled: make(chan struct{}),
		}
		client, err := minio.NewWithRegion("localhost:9000", "access", "secret", false, "us-east-1")
		Expect(err).To(BeNil())
		client.SetCustomTransport(transport)

		stager := &s3compat.S3Stager{Client: client, Bucket: "quarantine"}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err = stager.Check(ctx)
		Expect(err).To(HaveOccurred())
		Eventually(transport.requestStarted, time.Second).Should(BeClosed())
		Eventually(transport.requestCanceled, time.Second).Should(BeClosed())
	})
})
