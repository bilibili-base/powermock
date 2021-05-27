package pluginregistry

import (
	"context"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/interact"
)

// MockPlugin defines the Mock plugin
// It is used to generate interact.Response according to the given MockAPI_Response and interact.Request
type MockPlugin interface {
	Plugin
	MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error)
}
