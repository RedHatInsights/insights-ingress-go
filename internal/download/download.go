package download

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/stage/filebased"
)

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(cfg config.IngressConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		downloadID := chi.URLParam(r, "requestID")
		fileName, err := filebased.GetFileStorageName(downloadID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Invalid request ID"))
			return
		} else {
			filePath := filepath.Join(cfg.StorageConfig.StorageFileSystemPath, fileName)
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",
				fileName))
			w.Header().Set("Content-Type", "application/gzip")
			http.ServeFile(w, r, filePath)
		}
	}
}
