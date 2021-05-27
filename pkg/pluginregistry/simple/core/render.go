package core

import (
	"fmt"
	"strings"

	"github.com/brianvoe/gofakeit"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"

	"github.com/bilibili-base/powermock/pkg/interact"
)

// Context defines the context
type Context struct {
	Request *interact.Request `json:"request"`
}

// NewContext is used to create context
func NewContext(request *interact.Request) *Context {
	return &Context{
		Request: request,
	}
}

// RenderWithRequest is used to render $request... variable
func RenderWithRequest(ctx *Context, path string) string {
	data, err := jsoniter.Marshal(ctx.Request)
	if err != nil {
		return ""
	}
	return gjson.GetBytes(data, path).Str
}

// RenderWithRequest is used to render $mock... variable
func RenderWithFaker(ctx *Context, path string) string {
	switch path {
	case "name":
		return gofakeit.Name()
	case "url":
		return gofakeit.URL()
	case "lastname":
		return gofakeit.LastName()
	case "email":
		return gofakeit.Email()
	case "price":
		data := fmt.Sprint(gofakeit.Price(0, 10000))
		return data
	}
	return path
}

// SplitWithFirstSegment is used to extract the first segment of s divided by split
func SplitWithFirstSegment(s string, split string) (string, string) {
	i := strings.Index(s, split)
	if i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

// Render is used to render variables based on context
func Render(ctx *Context, path string) string {
	scope, subPath := SplitWithFirstSegment(path, ".")
	switch scope {
	case "$request":
		return RenderWithRequest(ctx, subPath)
	case "$mock":
		return RenderWithFaker(ctx, subPath)
	default:
		return path
	}
}

func init() {
	gofakeit.Seed(0)
}
