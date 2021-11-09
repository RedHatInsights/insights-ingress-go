package featureflags

import (
	"net/http"

	unleash "github.com/Unleash/unleash-client-go/v3"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

var Client *unleash.Client

func InitFFClient(cfg *config.IngressConfig) *unleash.Client {
	Client, err := unleash.NewClient(
		unleash.WithAppName("insights-ingress"),
		unleash.WithUrl(cfg.FeatureFlagsConfig.FFHostname + "/api/"),
		unleash.WithCustomHeaders(http.Header{"Authorization": {cfg.FeatureFlagsConfig.FFToken}}),
	)

	if err != nil {
		l.Log.Errorf("Error initializing feature flags client: %v", err)
	}

	return Client
}
