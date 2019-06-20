package version

import (
	"encoding/json"
	"net/http"

	cfg "github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
)

// GetVersion deals with calls to the version endpoint
func GetVersion(w http.ResponseWriter, r *http.Request) {
	v := &IngressVersion{
		Commit:  cfg.Get().OpenshiftBuildCommit,
		Version: cfg.Get().Version,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(v)
	if err != nil {
		l.Log.Error("Unable to get version")
		w.Write([]byte(`{"version": "unavailable"}`))
	} else {
		w.Write(jsonData)
	}
}
