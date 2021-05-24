package internal

import (
	"github.com/spf13/pflag"

	"github.com/storyicon/powermock/pkg/apimanager"
	grpcmockserver "github.com/storyicon/powermock/pkg/mockserver/grpc"
	httpmockserver "github.com/storyicon/powermock/pkg/mockserver/http"
	"github.com/storyicon/powermock/pkg/pluginregistry"
	"github.com/storyicon/powermock/pkg/util"
	"github.com/storyicon/powermock/pkg/util/logger"
)

// Config defines the powermock config with plugins
type Config struct {
	Log            *logger.Config
	GRPCMockServer *grpcmockserver.Config
	HTTPMockServer *httpmockserver.Config
	ApiManager     *apimanager.Config
	PluginRegistry *pluginregistry.Config
	Plugin         *PluginConfig
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Log:            logger.NewConfig(),
		GRPCMockServer: grpcmockserver.NewConfig(),
		HTTPMockServer: httpmockserver.NewConfig(),
		ApiManager:     apimanager.NewConfig(),
		PluginRegistry: pluginregistry.NewConfig(),
		Plugin:         NewPluginConfig(),
	}
}

// RegisterFlags is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.Log.RegisterFlagsWithPrefix(prefix, f)
	c.HTTPMockServer.RegisterFlagsWithPrefix(prefix, f)
	c.GRPCMockServer.RegisterFlagsWithPrefix(prefix, f)
	c.ApiManager.RegisterFlagsWithPrefix(prefix, f)
	c.PluginRegistry.RegisterFlagsWithPrefix(prefix, f)
	c.Plugin.RegisterFlagsWithPrefix(prefix, f)
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return util.CheckErrors(
		c.PluginRegistry.Validate(),
		c.GRPCMockServer.Validate(),
		c.HTTPMockServer.Validate(),
		c.ApiManager.Validate(),
		c.Plugin.Validate())
}
