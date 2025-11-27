package download_test

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/download"
)

var _ = Describe("Download", func() {
	var (
		configuration *config.IngressConfig
		tmpDir        string
		testFile      string
		router        *chi.Mux
		handler       http.HandlerFunc
		recorder      *httptest.ResponseRecorder
		request       *http.Request
	)

	BeforeEach(func() {
		var err error
		configuration = config.Get()
		tmpDir, err = os.MkdirTemp("", "testfiles")
		configuration.StorageConfig.StorageFileSystemPath = tmpDir
		Expect(err).NotTo(HaveOccurred())
		router = chi.NewRouter()

		handler = download.NewHandler(*configuration)
		router.Get("/download/{requestID}", handler)
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("should serve the file content", func() {
		testFile = filepath.Join(tmpDir, "1234.tar.gz")
		err := os.WriteFile(testFile, []byte("Test file content"), 0644)
		Expect(err).NotTo(HaveOccurred())

		request, err = http.NewRequest("GET", "/download/1234", nil)
		Expect(err).NotTo(HaveOccurred())

		// Create a response recorder
		recorder = httptest.NewRecorder()

		// Serve the request
		router.ServeHTTP(recorder, request)

		// Assert the response
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(Equal("Test file content"))
	})

	It("should return 404 if the file does not exist", func() {
		recorder = httptest.NewRecorder()
		request, _ = http.NewRequest("GET", "/download/99999", nil)

		router.ServeHTTP(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusNotFound))
	})

	It("should return 400 if the request ID is invalid", func() {
		recorder = httptest.NewRecorder()
		request, _ = http.NewRequest("GET", "/download/--", nil)

		router.ServeHTTP(recorder, request)
		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
	})
})
