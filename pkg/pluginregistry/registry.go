package pluginregistry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/storyicon/powermock/pkg/util/logger"
)

// Registry defines the management center of Plugins
type Registry interface {
	// MockPlugins is used to return all registered mock plugins
	MockPlugins() []MockPlugin
	// RegisterMockPlugins is used to register mock plugins
	RegisterMockPlugins(...MockPlugin) error

	// MockPlugins is used to return all registered match plugins
	MatchPlugins() []MatchPlugin
	// RegisterMatchPlugins is used to register match plugins
	RegisterMatchPlugins(...MatchPlugin) error

	// MockPlugins is used to return storage plugins
	StoragePlugins() []StoragePlugin
	// RegisterStoragePlugin is used to register storage plugins
	RegisterStoragePlugins(...StoragePlugin) error
}

// BasicRegistry is the basic implementation of pluginRegistry
type BasicRegistry struct {
	cfg *Config

	matchPlugins   []MatchPlugin
	mockPlugins    []MockPlugin
	storagePlugins []StoragePlugin
	registerer     prometheus.Registerer
	lock           sync.Mutex

	logger.Logger
}

// Config defines the config structure
type Config struct{}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (Registry, error) {
	service := &BasicRegistry{
		matchPlugins: []MatchPlugin{},
		mockPlugins:  []MockPlugin{},
		cfg:          cfg,
		registerer:   registerer,
		Logger:       logger.NewLogger("pluginsRegistry"),
	}
	return service, nil
}

// MockPlugins is used to return all registered mock plugins
func (b *BasicRegistry) MockPlugins() []MockPlugin {
	return b.mockPlugins
}

// RegisterMockPlugins is used to register mock plugins
func (b *BasicRegistry) RegisterMockPlugins(plugins ...MockPlugin) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.mockPlugins = append(b.mockPlugins, plugins...)
	return nil
}

// MockPlugins is used to return all registered match plugins
func (b *BasicRegistry) RegisterMatchPlugins(plugins ...MatchPlugin) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.matchPlugins = append(b.matchPlugins, plugins...)
	return nil
}

// RegisterMatchPlugins is used to register match plugins
func (b *BasicRegistry) MatchPlugins() []MatchPlugin {
	return b.matchPlugins
}

// MockPlugins is used to return storage plugins
func (b *BasicRegistry) StoragePlugins() []StoragePlugin {
	return b.storagePlugins
}

// RegisterStoragePlugin is used to register storage plugins
func (b *BasicRegistry) RegisterStoragePlugins(plugins ...StoragePlugin) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.storagePlugins = append(b.storagePlugins, plugins...)
	return nil
}
