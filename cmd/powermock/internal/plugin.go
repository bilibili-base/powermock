package internal

import (
	"github.com/spf13/pflag"

	pluginsgrpc "github.com/bilibili-base/powermock/pkg/pluginregistry/grpc"
	pluginshttp "github.com/bilibili-base/powermock/pkg/pluginregistry/http"
	pluginscript "github.com/bilibili-base/powermock/pkg/pluginregistry/script"
	pluginssimple "github.com/bilibili-base/powermock/pkg/pluginregistry/simple"
	pluginredis "github.com/bilibili-base/powermock/pkg/pluginregistry/storage/redis"
	"github.com/bilibili-base/powermock/pkg/util"
)

// PluginConfig defines the plugin config
type PluginConfig struct {
	Redis  *pluginredis.Config
	Simple *pluginssimple.Config
	GRPC   *pluginsgrpc.Config
	HTTP   *pluginshttp.Config
	Script *pluginscript.Config
}

// NewPluginConfig is used to create plugin config
func NewPluginConfig() *PluginConfig {
	return &PluginConfig{
		Redis:  pluginredis.NewConfig(),
		Simple: pluginssimple.NewConfig(),
		GRPC:   pluginsgrpc.NewConfig(),
		HTTP:   pluginshttp.NewConfig(),
		Script: pluginscript.NewConfig(),
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *PluginConfig) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.Redis.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.Simple.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.GRPC.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.HTTP.RegisterFlagsWithPrefix(prefix+"plugin.", f)
	c.Script.RegisterFlagsWithPrefix(prefix+"plugin.", f)
}

// Validate is used to validate config and returns error on failure
func (c *PluginConfig) Validate() error {
	return util.ValidateConfigs(
		c.Redis,
		c.Simple,
		c.GRPC,
		c.HTTP,
		c.Script,
	)
}
