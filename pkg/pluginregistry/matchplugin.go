package pluginregistry

import (
	"context"

	"github.com/bilibili-base/powermock/apis/v1alpha1"
	"github.com/bilibili-base/powermock/pkg/interact"
)

// MatchPlugin defines matching plugins
// It is used to determine whether interact.Request satisfies the matching condition of MockAPI_Condition
type MatchPlugin interface {
	Plugin
	Match(ctx context.Context, request *interact.Request, condition *v1alpha1.MockAPI_Condition) (match bool, err error)
}
