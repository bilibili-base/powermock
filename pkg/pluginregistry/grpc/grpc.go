package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/hashicorp/go-multierror"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/interact"
	"github.com/storyicon/powermock/pkg/pluginregistry"
	"github.com/storyicon/powermock/pkg/util/logger"
)

var _ pluginregistry.MockPlugin = &Plugin{}

// MethodDescGetter is used to obtain the MethodDescriptor corresponding to a given method
type MethodDescGetter func(method string) (*desc.MethodDescriptor, bool)

// Plugin implements Mock for gRPC request
type Plugin struct {
	cfg *Config

	methodDescGetter MethodDescGetter
	registerer       prometheus.Registerer
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
func New(cfg *Config, methodDescGetter MethodDescGetter, logger logger.Logger, registerer prometheus.Registerer) (*Plugin, error) {
	if methodDescGetter == nil {
		return nil, errors.New("method descriptor getter is required")
	}
	service := &Plugin{
		cfg:              cfg,
		methodDescGetter: methodDescGetter,
		registerer:       registerer,
		Logger:           logger.NewLogger("gRPCPlugin"),
	}
	return service, nil
}

// Name is used to return the plugin name
func (s *Plugin) Name() string {
	return "grpc"
}

// MockResponse is used to generate interact.Response according to the given MockAPI_Response and interact.Request
func (s *Plugin) MockResponse(ctx context.Context, mock *v1alpha1.MockAPI_Response, request *interact.Request, response *interact.Response) (abort bool, err error) {
	if request.Protocol != interact.ProtocolGRPC {
		return false, nil
	}
	md, ok := s.methodDescGetter(request.Path)
	if !ok {
		return true, fmt.Errorf("unable to find descriptor: %s", request.Path)
	}
	data := response.Body.Bytes()
	message := dynamic.NewMessage(md.GetOutputType())
	if err := message.UnmarshalJSONPB(&jsonpb.Unmarshaler{}, data); err != nil {
		return true, multierror.Prefix(err, "failed to unmarshal:")
	}
	binaryData, err := message.Marshal()
	if err != nil {
		return true, multierror.Prefix(err, "failed to marshal:")
	}
	response.Body = interact.NewBytesMessage(binaryData)
	return false, nil
}
