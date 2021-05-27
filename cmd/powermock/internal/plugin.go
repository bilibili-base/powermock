// Copyright 2021 bilibili-base
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
