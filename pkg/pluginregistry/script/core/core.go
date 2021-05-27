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

package core

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"rogchap.com/v8go"

	"github.com/bilibili-base/powermock/pkg/interact"
)

// MatchRequestByJavascript is used to match mock case by javascript
func MatchRequestByJavascript(ctx context.Context, request *interact.Request, script string) (bool, error) {
	vm, err := v8go.NewContext()
	if err != nil {
		return false, err
	}
	requestRaw, err := jsoniter.MarshalToString(request)
	if err != nil {
		return false, err
	}
	_, err = RunScript(ctx, vm, fmt.Sprintf("const request = %s", requestRaw))
	if err != nil {
		return false, err
	}
	value, err := RunScript(ctx, vm, script)
	if err != nil {
		return false, err
	}
	return value.Boolean(), nil
}

// MockResponseByJavascript is used to mock response by javascript
func MockResponseByJavascript(ctx context.Context, request *interact.Request, response *interact.Response, script string) error {
	vm, err := v8go.NewContext()
	if err != nil {
		return err
	}
	requestRaw, err := jsoniter.MarshalToString(request)
	if err != nil {
		return err
	}
	_, err = RunScript(ctx, vm, fmt.Sprintf("const request = %s", requestRaw))
	if err != nil {
		return err
	}
	value, err := RunScript(ctx, vm, script)
	if err != nil {
		return err
	}
	responseRaw, err := value.MarshalJSON()
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(responseRaw, &response)
	if err != nil {
		return err
	}
	return nil
}

// RunScript is used to run javascript with context
func RunScript(ctx context.Context, v8Context *v8go.Context, script string) (*v8go.Value, error) {
	valCh := make(chan *v8go.Value, 1)
	errCh := make(chan error, 1)
	go func() {
		val, err := v8Context.RunScript(script, "main.js")
		if err != nil {
			errCh <- err
			return
		}
		valCh <- val
	}()
	var terminateFunc = func() error {
		vm, _ := v8Context.Isolate()
		vm.TerminateExecution()
		return <-errCh
	}
	select {
	case val := <-valCh:
		return val, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, terminateFunc()
	}
}
