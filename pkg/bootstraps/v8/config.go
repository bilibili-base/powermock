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

package bootstrap

import (
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/apimanager"
	grpcmockserver "github.com/bilibili-base/powermock/pkg/mockserver/grpc"
	httpmockserver "github.com/bilibili-base/powermock/pkg/mockserver/http"
	"github.com/bilibili-base/powermock/pkg/pluginregistry"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Config defines the powermock config with plugins
type Config struct {
	MetricsAddress string
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
		MetricsAddress: "0.0.0.0:8081",
		Log:            logger.NewConfig(),
		GRPCMockServer: grpcmockserver.NewConfig(),
		HTTPMockServer: httpmockserver.NewConfig(),
		ApiManager:     apimanager.NewConfig(),
		PluginRegistry: pluginregistry.NewConfig(),
		Plugin:         NewPluginConfig(),
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.MetricsAddress, prefix+"metrics.address", c.MetricsAddress, "metrics address")
	c.Log.RegisterFlagsWithPrefix(prefix, f)
	c.HTTPMockServer.RegisterFlagsWithPrefix(prefix, f)
	c.GRPCMockServer.RegisterFlagsWithPrefix(prefix, f)
	c.ApiManager.RegisterFlagsWithPrefix(prefix, f)
	c.PluginRegistry.RegisterFlagsWithPrefix(prefix, f)
	c.Plugin.RegisterFlagsWithPrefix(prefix, f)
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return util.ValidateConfigs(
		c.PluginRegistry,
		c.GRPCMockServer,
		c.HTTPMockServer,
		c.ApiManager,
		c.Plugin,
	)
}
