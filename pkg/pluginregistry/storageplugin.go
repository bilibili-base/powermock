package pluginregistry

import (
	"context"
)

// StoragePlugin defines the storage plugin
type StoragePlugin interface {
	Plugin
	// Set is used to set key-val pair
	Set(ctx context.Context, key string, val string) error
	// Delete is used to delete specified key
	Delete(ctx context.Context, key string) error
	// List is used to list all key-val pairs in storage
	List(ctx context.Context) (map[string]string, error)
}
