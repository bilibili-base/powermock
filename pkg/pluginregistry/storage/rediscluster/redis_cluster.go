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

package rediscluster

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/pluginregistry"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

var _ pluginregistry.StoragePlugin = &Plugin{}

// Plugin defines a storage plugin with Redis as the backend
type Plugin struct {
	cfg *Config

	client     *redis.ClusterClient
	registerer prometheus.Registerer
	logger.Logger

	announcement chan struct{}
}

// Config defines the config structure
type Config struct {
	Enable    bool
	Addresses []string
	Password  string
	Prefix    string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: false,
		Addresses: []string{
			"127.0.0.1:6379",
		},
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
	f.BoolVar(&c.Enable, prefix+"rediscluster.enable", c.Enable, "define whether the component is enabled")
	f.StringSliceVar(&c.Addresses, prefix+"rediscluster.addr", c.Addresses, "redis address(ip:port format), support multiple calls to form a redis cluster")
	f.StringVar(&c.Password, prefix+"rediscluster.password", c.Password, "redis password")
	f.StringVar(&c.Prefix, prefix+"rediscluster.prefix", c.Prefix, "storage prefix")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if len(c.Addresses) == 0 {
		return errors.New("redis address is required")
	}
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (pluginregistry.StoragePlugin, error) {
	s := &Plugin{
		cfg:          cfg,
		registerer:   registerer,
		Logger:       logger.NewLogger("redisPlugin"),
		announcement: make(chan struct{}, 1),
	}
	s.LogInfo(nil, "start to init redis(addr: %s)", cfg.Addresses)
	if err := s.initRedis(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Plugin) initRedis() error {
	s.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    s.cfg.Addresses,
		Password: s.cfg.Password,
	})
	if err := s.client.Ping(context.Background()).Err(); err != nil {
		return err
	}
	return nil
}

// Start is used to start the plugin
func (s *Plugin) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	return s.watch(ctx, cancelFunc)
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "redis"
}

// Set is used to set key-val pair
func (s *Plugin) Set(ctx context.Context, key string, val string) error {
	actualKey := s.cfg.Prefix + key
	s.LogInfo(nil, "redis storage set key: %s", actualKey)
	if err := s.client.Set(ctx, actualKey, val, 0).Err(); err != nil {
		return err
	}
	return s.incrRevision(ctx)
}

// Delete is used to delete specified key
func (s *Plugin) Delete(ctx context.Context, key string) error {
	actualKey := s.cfg.Prefix + key
	s.LogInfo(nil, "redis storage delete key: %s", actualKey)
	if err := s.client.Del(ctx, actualKey).Err(); err != nil {
		return err
	}
	return s.incrRevision(ctx)
}

// List is used to list all key-val pairs in storage
func (s *Plugin) List(ctx context.Context) (map[string]string, error) {
	wildcardKey := s.cfg.Prefix + "*"
	data := map[string]string{}
	var lock sync.Mutex
	err := s.client.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		s.LogInfo(nil, "starting fetch keys from %s", client.Options().Addr)
		iterator := client.Scan(ctx, 0, wildcardKey, 10000).Iterator()
		for iterator.Next(ctx) {
			key := iterator.Val()
			if GetRevisionKey(s.cfg.Prefix) == key {
				continue
			}
			value, err := client.Get(ctx, key).Result()
			if err != nil {
				return err
			}
			key = strings.TrimPrefix(key, s.cfg.Prefix)
			lock.Lock()
			data[key] = value
			lock.Unlock()
		}
		if err := iterator.Err(); err != nil {
			if err == redis.Nil {
				return nil
			}
			return err
		}
		return nil
	})
	s.LogInfo(nil, "total %d keys loaded", len(data))
	if err != nil {
		return nil, err
	}
	return data, nil
}

// //////////// Simple Watch Implement ///////////////

func (s *Plugin) incrRevision(ctx context.Context) error {
	return s.client.Incr(ctx, GetRevisionKey(s.cfg.Prefix)).Err()
}

func (s *Plugin) getRevision(ctx context.Context) (int64, error) {
	return s.client.Get(ctx, GetRevisionKey(s.cfg.Prefix)).Int64()
}

// GetAnnouncement is used to get announcement
func (s *Plugin) GetAnnouncement() chan struct{} {
	return s.announcement
}

func (s *Plugin) watch(ctx context.Context, cancelFunc context.CancelFunc) error {
	revision, err := s.getRevision(ctx)
	if err != nil && err != redis.Nil {
		return err
	}
	s.LogInfo(nil, "start to watch redis revisions, current: %d", revision)
	util.StartServiceAsync(ctx, cancelFunc, s.Logger, func() error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				got, err := s.getRevision(ctx)
				if err != nil && err != redis.Nil {
					s.LogError(nil, "failed to get revision key: %s", err)
					continue
				}
				if revision != got {
					s.LogInfo(nil, "event found (known %d vs got %d)", revision, got)
					revision = got
					select {
					case s.announcement <- struct{}{}:
					default:
					}
				}
			case <-ctx.Done():
				s.LogWarn(nil, "redis stop watching...")
				return nil
			}
		}
	}, func() error {
		return nil
	})
	return nil
}

// GetRevisionKey is used to get revision key
func GetRevisionKey(prefix string) string {
	return prefix + "__REVISION__"
}
