package download

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	"github.com/redhatinsights/insights-ingress-go/internal/stage/filebased"
)

// NewHandler returns a http handler configured with a Pipeline
func NewHandler(cfg config.IngressConfig) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		downloadID := chi.URLParam(r, "requestID")
		fileName, filePath, err := filebased.GetFileStorageName(downloadID, cfg.StorageConfig.StorageFileSystemPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request: Invalid request ID"))
			return
		} else {
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",
				fileName))
			w.Header().Set("Content-Type", "application/gzip")
			http.ServeFile(w, r, filePath)
		}
	}
}
