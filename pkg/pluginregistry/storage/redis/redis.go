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

package redis

import (
	"context"
	"errors"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Plugin defines a storage plugin with Redis as the backend
type Plugin struct {
	cfg *Config

	client     *redis.Client
	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable   bool
	Addr     string
	Password string
	DB       int
	Prefix   string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: false,
		Addr:   "127.0.0.1:6379",
		Prefix: "/powermock/",
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
	f.StringVar(&c.Addr, prefix+"redis.addr", c.Addr, "redis address(ip:port format)")
	f.StringVar(&c.Password, prefix+"redis.password", c.Password, "redis password")
	f.IntVar(&c.DB, prefix+"redis.db", c.DB, "redis database to use")
	f.StringVar(&c.Prefix, prefix+"redis.prefix", c.Prefix, "storage prefix")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.Addr == "" {
		return errors.New("redis address is required")
	}
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (*Plugin, error) {
	s := &Plugin{
		cfg:        cfg,
		registerer: registerer,
		Logger:     logger.NewLogger("redisPlugin"),
	}
	s.LogInfo(nil, "start to init redis(addr: %s)", cfg.Addr)
	if err := s.initRedis(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Plugin) initRedis() error {
	s.client = redis.NewClient(&redis.Options{
		Addr:     s.cfg.Addr,
		Password: s.cfg.Password,
		DB:       s.cfg.DB,
	})
	if err := s.client.Ping(context.Background()).Err(); err != nil {
		return err
	}
	return nil
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "redis"
}

// Set is used to set key-val pair
func (s *Plugin) Set(ctx context.Context, key string, val string) error {
	actualKey := s.cfg.Prefix + key
	s.LogInfo(nil, "redis storage set key: %s", actualKey)
	return s.client.Set(ctx, actualKey, val, 0).Err()
}

// Delete is used to delete specified key
func (s *Plugin) Delete(ctx context.Context, key string) error {
	actualKey := s.cfg.Prefix + key
	s.LogInfo(nil, "redis storage delete key: %s", actualKey)
	return s.client.Del(ctx, actualKey).Err()
}

// List is used to list all key-val pairs in storage
func (s *Plugin) List(ctx context.Context) (map[string]string, error) {
	wildcardKey := s.cfg.Prefix + "*"
	iterator := s.client.Scan(ctx, 0, wildcardKey, 10000).Iterator()
	var keys []string
	for iterator.Next(ctx) {
		key := iterator.Val()
		keys = append(keys, key)
	}
	if err := iterator.Err(); err != nil {
		if err == redis.Nil {
			return map[string]string{}, nil
		}
		return nil, err
	}
	if len(keys) == 0 {
		return map[string]string{}, nil
	}

	s.LogInfo(nil, "redis storage start to MGet keys, total %d", len(keys))
	values, err := s.client.MGet(ctx, keys...).Result()
	s.LogInfo(nil, "redis storage MGet keys finished")
	if err != nil {
		return nil, err
	}
	data := make(map[string]string, len(keys))
	for i := 0; i < len(keys); i++ {
		key := strings.TrimPrefix(keys[i], s.cfg.Prefix)
		value := values[i].(string)
		data[key] = value
	}
	return data, nil
}
