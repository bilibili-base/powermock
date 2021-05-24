package core

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"rogchap.com/v8go"

	"github.com/storyicon/powermock/pkg/interact"
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
	_, err = vm.RunScript(fmt.Sprintf("const request = %s", requestRaw), "main.js")
	if err != nil {
		return false, err
	}
	value, err := vm.RunScript(script, "main.js")
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
	_, err = vm.RunScript(fmt.Sprintf("const request = %s", requestRaw), "main.js")
	if err != nil {
		return err
	}
	value, err := vm.RunScript(script, "main.js")
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