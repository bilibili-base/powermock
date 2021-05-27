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

package util

import (
	"context"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// StartServiceAsync is used to start service async
func StartServiceAsync(ctx context.Context, cancelFunc context.CancelFunc, logger logger.Logger, serveFn func() error, stopFn func() error) {
	if serveFn == nil {
		return
	}
	go func() {
		logger.LogInfo(nil, "starting service")
		go func() {
			if err := serveFn(); err != nil {
				logger.LogError(nil, "error serving service: %s", err)
			}
			if cancelFunc != nil {
				cancelFunc()
			}
		}()
		<-ctx.Done()
		logger.LogInfo(nil, "stopping service")
		if stopFn() != nil {
			logger.LogInfo(nil, "stopping service gracefully")
			if err := stopFn(); err != nil {
				logger.LogWarn(nil, "error occurred while stopping service: %s", err)
			}
		}
		logger.LogInfo(nil, "exiting service")
	}()
}
