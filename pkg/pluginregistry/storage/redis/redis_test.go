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
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/bilibili-base/powermock/pkg/util/logger"
)

func TestPlugin(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute)
	defer cancel()
	plugin, err := New(NewConfig(), logger.NewDefault("redis"), prometheus.DefaultRegisterer)
	assert.Equal(t, nil, err)
	assert.Equal(t, nil, plugin.Start(ctx, cancel))

	total := 3
	go func() {
		for i := 0; i < total; i++ {
			assert.Equal(t, nil, plugin.Set(ctx, "a", "1"))
			time.Sleep(time.Second * 2)
		}
	}()
	count := 0
	announcement := plugin.GetAnnouncement()
Loop:
	for {
		select {
		case <-announcement:
			count++
			if total == count {
				break Loop
			}
		case <-ctx.Done():
			assert.Fail(t, "event lost")
		}
	}
}
