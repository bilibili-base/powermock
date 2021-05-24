package internal

import (
	"github.com/spf13/pflag"

	pluginsgrpc "github.com/storyicon/powermock/pkg/pluginregistry/grpc"
	pluginshttp "github.com/storyicon/powermock/pkg/pluginregistry/http"
	pluginssimple "github.com/storyicon/powermock/pkg/pluginregistry/simple"
	pluginredis "github.com/storyicon/powermock/pkg/pluginregistry/storage/redis"
	"github.com/storyicon/powermock/pkg/util"
)

// PluginConfig defines the plugin config
type PluginConfig struct {
	Redis  *pluginredis.Config
	Simple *pluginssimple.Config
	GRPC   *pluginsgrpc.Config
	HTTP   *pluginshttp.Config
}

// NewPluginConfig is used to create plugin config
func NewPluginConfig() *PluginConfig {
	return &PluginConfig{
		Redis:  pluginredis.NewConfig(),
		Simple: pluginssimple.NewConfig(),
		GRPC:   pluginsgrpc.NewConfig(),
		HTTP:   pluginshttp.NewConfig(),
	}
}

// RegisterFlags is used to register flags
func (c *PluginConfig) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.Redis.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.Simple.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.GRPC.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.HTTP.RegisterFlagsWithPrefix(prefix+"plugin.", f)
}

// Validate is used to validate config and returns error on failure
func (c *PluginConfig) Validate() error {
	return util.CheckErrors(
		c.Redis.Validate(),
		c.Simple.Validate(),
		c.GRPC.Validate(),
		c.HTTP.Validate(),
	)
}
