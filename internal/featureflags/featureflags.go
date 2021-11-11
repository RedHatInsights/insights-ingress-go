package featureflags

import (
	"fmt"
	"errors"

	unleash "github.com/Unleash/unleash-client-go/v3"
	"github.com/redhatinsights/insights-ingress-go/internal/config"
	l "github.com/redhatinsights/insights-ingress-go/internal/logger"
)

type FeatureFlagClient interface {
	InitializeClient() error
	CheckFlag(flag string) bool
}

func NewFeatureFlagClient(impl string, cfg *config.IngressConfig) (FeatureFlagClient, error) {

	switch impl {
	case "fake":
		configuredClient := FakeFeatureFlagClient{
		}

		return &configuredClient, nil
	case "unleash":
		configuredClient := UnleashFeatureFlagClient{
			Config: cfg,
		}
			
		return &configuredClient, nil
	default:
		return nil, errors.New("Invalid Feature Flag Client impl requested")
	}
}

type UnleashFeatureFlagClient struct {
	Config *config.IngressConfig
	Client *unleash.Client
}

func (ufc *UnleashFeatureFlagClient) InitializeClient() error {
	url := fmt.Sprintf("%s://%s:%v",
				ufc.Config.FeatureFlagsConfig.FFScheme,
				ufc.Config.FeatureFlagsConfig.FFHostname,
				ufc.Config.FeatureFlagsConfig.FFPort,
			   )	
	unleashClient, err := unleash.NewClient(
		unleash.WithAppName("insights-ingress"),
		unleash.WithUrl(url +  "/api/"),
		unleash.WithListener(&unleash.DebugListener{}),
	//	unleash.WithCustomHeaders(http.Header{"Authorization": {ufc.Config.FeatureFlagsConfig.FFToken}}),
	)
	if err != nil {
		l.Log.Errorf("Error initializing feature flags client: %v", err)
		return err
	}
	ufc.Client = unleashClient
	return nil

}

func (ufc *UnleashFeatureFlagClient) CheckFlag(flag string) bool {
	result := ufc.Client.IsEnabled(flag)
	return result
}

type FakeFeatureFlagClient struct {
}

func (ffc *FakeFeatureFlagClient) InitializeClient() error {
	l.Log.Info("Initialized fake client")
	return nil
}

func (ffc *FakeFeatureFlagClient) CheckFlag(flag string) bool {
	return false
}

