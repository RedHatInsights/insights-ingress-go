package version

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	cfg "github.com/redhatinsights/insights-ingress-go/config"
	l "github.com/redhatinsights/insights-ingress-go/logger"
	"go.uber.org/zap"
)

// GetVersion deals with calls to the version endpoint
func GetVersion(w http.ResponseWriter, r *http.Request) {
	v := &IngressVersion{
		Commit:  cfg.Get().OpenshiftBuildCommit,
		Version: ReadVersion(),
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

// ReadVersion gets the version from a version file
func ReadVersion() string {
	var dat []byte
	dat, err := ioutil.ReadFile("/tmp/src/VERSION")
	if err != nil {
		l.Log.Error("Unable to read version", zap.Error(err))
		dat = []byte("0.0.0")
	}
	return string(dat)
}
