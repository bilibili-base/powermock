package pluginregistry

import (
	"context"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/interact"
)

// Plugin defines the most basic requirements for plug-ins
// All plugins should implement this interface
type Plugin interface {
	Name() string
}

// MatchPlugin defines matching plugins
// It is used to determine whether interact.Request satisfies the matching condition of MockAPI_Condition
type MatchPlugin interface {
	Plugin
	Match(ctx context.Context, request *interact.Request, condition *v1alpha1.MockAPI_Condition) (match bool, err error)
}

// MockPlugin defines the Mock plugin
// It is used to generate interact.Response according to the given MockAPI_Response and interact.Request
type MockPlugin interface {
	Plugin
	// TODO: Consider merging 'abort' and 'err' in the return value
	MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error)
}

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






