package track

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"github.com/redhatinsights/platform-go-middlewares/identity"

	"github.com/sirupsen/logrus"
)

// NewHandlers returns an http handler for tracking
func NewHandler(
	cfg config.IngressConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var id identity.XRHID
		reqID := chi.URLParam(r, "requestID")
		requestLogger := l.Log.WithFields(logrus.Fields{"source_host": cfg.Hostname, "name": "ingress"})

		logerr := func(msg string, err error) {
			requestLogger.WithFields(logrus.Fields{"error": err}).Error(msg)
		}

		if cfg.Auth == true {
			id = identity.Get(r.Context())
		}

		response, err := http.Get(cfg.PayloadTrackerURL + reqID)
		if err != nil {
			logerr("Failed to get payload status", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logerr("Unable to read response body", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var pt TrackerResponse
		if err = json.Unmarshal(body, &pt); err != nil {
			logerr("Failed to unmarshal tracker json", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(pt.Data) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if pt.Data[0].Account != id.Identity.AccountNumber {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(body)

	}
}