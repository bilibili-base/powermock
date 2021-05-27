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

package script

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/apis/v1alpha1"
	"github.com/bilibili-base/powermock/pkg/interact"
	"github.com/bilibili-base/powermock/pkg/pluginregistry"
	"github.com/bilibili-base/powermock/pkg/pluginregistry/script/core"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

var _ pluginregistry.MockPlugin = &Plugin{}
var _ pluginregistry.MatchPlugin = &Plugin{}

// Plugin implements Mock for http request
type Plugin struct {
	cfg *Config

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable bool
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: true,
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return c.Enable
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.BoolVar(&c.Enable, prefix+"script.enable", c.Enable, "define whether the component is enabled")
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
	return "script"
}

// Match is used to determine whether interact.Request satisfies the matching condition of MockAPI_Condition
func (s *Plugin) Match(ctx context.Context, request *interact.Request, condition *v1alpha1.MockAPI_Condition) (match bool, err error) {
	script := condition.GetScript()
	if script == nil {
		return false, nil
	}
	if script.Lang != "javascript" {
		return false, fmt.Errorf("script language %s is not supported yet", script.Lang)
	}
	return core.MatchRequestByJavascript(ctx, request, script.GetContent())
}

// MockResponse is used to generate interact.Response according to the given MockAPI_Response and interact.Request
func (s *Plugin) MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error) {
	script := mock.GetScript()
	if script == nil {
		return false, nil
	}
	if script.Lang != "javascript" {
		return true, fmt.Errorf("script language %s is not supported yet", script.Lang)
	}

	// get timeout
	timeout := time.Second
	if t := script.GetTimeout(); t != nil {
		milliseconds := t.AsDuration().Milliseconds()
		if milliseconds > 0 && milliseconds < 3000 {
			timeout = t.AsDuration()
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err = core.MockResponseByJavascript(ctx, request, response, script.GetContent())
	if err != nil {
		return true, err
	}
	return false, nil
}
