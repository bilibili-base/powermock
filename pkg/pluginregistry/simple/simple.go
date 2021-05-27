package simple

import (
	"context"
	"io"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"github.com/valyala/fasttemplate"

	"github.com/storyicon/powermock/pkg/pluginregistry"
	"github.com/storyicon/powermock/pkg/pluginregistry/simple/core"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/interact"
	"github.com/storyicon/powermock/pkg/util/logger"
)

var (
	_ pluginregistry.MockPlugin  = &Plugin{}
	_ pluginregistry.MatchPlugin = &Plugin{}
)

// Plugin defines the most basic matching and Mock plug-ins
type Plugin struct {
	cfg *Config

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable bool
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable: true,
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return true
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (*Plugin, error) {
	service := &Plugin{
		cfg:        cfg,
		registerer: registerer,
		Logger:     logger.NewLogger("simplePlugin"),
	}
	return service, nil
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "simple"
}

// Match is used to determine whether interact.Request satisfies the matching condition of MockAPI_Condition
func (s *Plugin) Match(ctx context.Context, request *interact.Request, condition *v1alpha1.MockAPI_Condition) (match bool, err error) {
	simple := condition.GetSimple()
	if simple == nil {
		return false, nil
	}
	c := core.NewContext(request)
	for _, item := range simple.Items {
		operandX := core.Render(c, item.OperandX)
		operandY := core.Render(c, item.OperandY)
		matched, err := core.Match(operandX, item.Operator, operandY)
		if err != nil {
			return false, err
		}
		if item.Opposite {
			matched = !matched
		}
		if matched {
			if simple.UseOrAmongItems {
				return true, nil
			}
			continue
		} else {
			if simple.UseOrAmongItems {
				continue
			}
			return false, nil
		}
	}
	return true, nil
}

// MockResponse is used to generate interact.Response according to the given MockAPI_Response and interact.Request
func (s *Plugin) MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error) {
	simple := mock.GetSimple()
	if simple == nil {
		return false, nil
	}
	c := core.NewContext(request)

	// Render Code
	response.Code = simple.GetCode()
	// Render Headers
	response.Header = map[string]string{}
	for key, val := range simple.GetHeader() {
		response.Header[key] = core.Render(c, val)
	}
	// Render Trailers
	response.Trailer = map[string]string{}
	for key, val := range simple.GetTrailer() {
		response.Trailer[key] = val
	}
	// Render Body
	data, err := fasttemplate.ExecuteFuncStringWithErr(simple.GetBody(), "{{", "}}", func(w io.Writer, tag string) (int, error) {
		return w.Write([]byte(core.Render(c, strings.TrimSpace(tag))))
	})
	if err != nil {
		return true, err
	}
	response.Body = interact.NewBytesMessage([]byte(data))
	return false, nil
}
