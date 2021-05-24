package http

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/interact"
	"github.com/storyicon/powermock/pkg/util/logger"
)

// Plugin implements Mock for http request
type Plugin struct {
	cfg *Config

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{}
}

// RegisterFlags is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {

}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (*Plugin, error) {
	service := &Plugin{
		cfg:        cfg,
		registerer: registerer,
		Logger:     logger.NewLogger("httpPlugin"),
	}
	return service, nil
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "http"
}

// MockResponse is used to generate interact.Response according to the given MockAPI_Response and interact.Request
func (s *Plugin) MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error) {
	if request.Protocol != interact.ProtocolHTTP {
		return false, nil
	}
	response.Header["content-type"] = "application/json"
	return false, nil
}
