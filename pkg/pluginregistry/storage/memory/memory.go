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

package memory

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/pluginregistry"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Plugin defines a storage plugin with Redis as the backend
type Plugin struct {
	cfg *Config

	data         sync.Map
	announcement chan struct{}
	registerer   prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable bool
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: false,
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return c.Enable
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.BoolVar(&c.Enable, prefix+"redis.enable", c.Enable, "define whether the component is enabled")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (pluginregistry.StoragePlugin, error) {
	s := &Plugin{
		cfg:          cfg,
		registerer:   registerer,
		Logger:       logger.NewLogger("redisPlugin"),
		announcement: make(chan struct{}, 10),
	}
	return s, nil
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "memory"
}

// Set is used to set key-val pair
func (s *Plugin) Set(ctx context.Context, key string, val string) error {
	s.data.Store(key, val)
	s.announcement <- struct{}{}
	return nil
}

// Start is used to start the plugin
func (s *Plugin) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	return nil
}

// Delete is used to delete specified key
func (s *Plugin) Delete(ctx context.Context, key string) error {
	s.data.Delete(key)
	s.announcement <- struct{}{}
	return nil
}

// List is used to list all key-val pairs in storage
func (s *Plugin) List(ctx context.Context) (map[string]string, error) {
	data := make(map[string]string)
	s.data.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(string)
		return true
	})
	return data, nil
}

func (s *Plugin) GetAnnouncement() chan struct{} {
	return s.announcement
}
