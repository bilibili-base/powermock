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
