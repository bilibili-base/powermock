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
