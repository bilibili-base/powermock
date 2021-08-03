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

package grpc

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bilibili-base/powermock/pkg/apimanager"
	"github.com/bilibili-base/powermock/pkg/interact"
	"github.com/bilibili-base/powermock/pkg/protomanager"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Provider defines the mock server interface
type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc) error
	GetProtoManager() protomanager.Provider
}

// MockServer is the implement of gRPC mock server
type MockServer struct {
	cfg *Config

	protoManager protomanager.Provider
	apiManager   apimanager.Provider

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable       bool
	Address      string
	ProtoManager *protomanager.Config
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable:       true,
		Address:      "0.0.0.0:30002",
		ProtoManager: protomanager.NewConfig(),
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return c.Enable
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	c.ProtoManager.RegisterFlagsWithPrefix(prefix+"gRPCMockServer.", f)
	f.BoolVar(&c.Enable, prefix+"gRPCMockServer.enable", c.Enable, "define whether the component is enabled")
	f.StringVar(&c.Address, prefix+"gRPCMockServer.address", c.Address, "address to listen")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("the address of mockserver is required")
	}
	return util.CheckErrors(c.ProtoManager.Validate())
}

// New is used to init service
func New(cfg *Config,
	apiManager apimanager.Provider,
	logger logger.Logger, registerer prometheus.Registerer) (Provider, error) {
	s := &MockServer{
		cfg:        cfg,
		apiManager: apiManager,
		registerer: registerer,
		Logger:     logger.NewLogger("gRPCMockServer"),
	}
	if err := s.setup(); err != nil {
		return nil, err
	}
	return s, nil
}

// GetProtoManager is used to get proto manager
func (s *MockServer) GetProtoManager() protomanager.Provider {
	return s.protoManager
}

// Start is used to start the service
func (s *MockServer) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	s.LogInfo(nil, "stating proto manager")
	if err := s.protoManager.Start(ctx, cancelFunc); err != nil {
		return err
	}

	s.LogInfo(nil, "starting gRPC mock server on: %s", s.cfg.Address)
	server := grpc.NewServer(grpc.UnknownServiceHandler(s.handleStream))
	listener, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	util.StartServiceAsync(ctx, cancelFunc, s.Logger.NewLogger("gRPC"), func() error {
		return server.Serve(listener)
	}, func() error {
		server.GracefulStop()
		return nil
	})
	return nil
}

func (s *MockServer) setup() error {
	if err := s.setupProtoManager(); err != nil {
		return err
	}
	return nil
}

func (s *MockServer) setupProtoManager() error {
	service, err := protomanager.New(s.cfg.ProtoManager, s.Logger, s.registerer)
	if err != nil {
		return err
	}
	s.protoManager = service
	return nil
}

func (s *MockServer) handleStream(srv interface{}, stream grpc.ServerStream) error {
	fullMethodName, ok := grpc.MethodFromServerStream(stream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	md, _ := metadata.FromIncomingContext(stream.Context())
	s.LogInfo(map[string]interface{}{
		"path":     fullMethodName,
		"metadata": md,
	}, "request received")

	method, ok := s.protoManager.GetMethod(fullMethodName)
	if !ok {
		return status.Errorf(codes.NotFound, "method not found")
	}
	request := dynamic.NewMessage(method.GetInputType())
	if err := stream.RecvMsg(request); err != nil {
		return status.Errorf(codes.Unknown, "failed to recv request")
	}
	data, err := request.MarshalJSONPB(&jsonpb.Marshaler{})
	if err != nil {
		return status.Errorf(codes.Unknown, "failed to marshal request")
	}
	response, err := s.apiManager.MockResponse(context.TODO(), &interact.Request{
		Protocol: interact.ProtocolGRPC,
		Method:   http.MethodPost,
		Host:     getAuthorityFromMetadata(md),
		Path:     fullMethodName,
		Header:   getHeadersFromMetadata(md),
		Body:     interact.NewBytesMessage(data),
	})
	if err != nil {
		return err
	}

	stream.SetTrailer(metadata.New(response.Trailer))
	if err := stream.SetHeader(metadata.New(response.Header)); err != nil {
		return status.Errorf(codes.Unavailable, "failed to set header: %s", err)
	}
	if response.Code != 0 {
		return status.Errorf(codes.Code(response.Code), "expected code is: %d", response.Code)
	}
	if err := stream.SendMsg(response.Body); err != nil {
		return status.Errorf(codes.Internal, "failed to send message: %s", err)
	}
	return nil
}

// getHeadersFromMetadata is used to convert Metadata to Headers
func getHeadersFromMetadata(md metadata.MD) map[string]string {
	headers := map[string]string{}
	for key, values := range md {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// getAuthorityFromMetadata is used to get authority from metadata
func getAuthorityFromMetadata(md metadata.MD) string {
	if md != nil {
		values := md[":authority"]
		if len(values) != 0 {
			return values[0]
		}
	}
	return ""
}
