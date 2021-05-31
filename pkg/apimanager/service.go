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

package apimanager

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bilibili-base/powermock/apis/v1alpha1"
	"github.com/bilibili-base/powermock/pkg/interact"
	"github.com/bilibili-base/powermock/pkg/pluginregistry"
	"github.com/bilibili-base/powermock/pkg/pluginregistry/storage/memory"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Provider defines the APIManager interface
// It is used to manage MockAPI, plug-ins, and generate MockResponse
type Provider interface {
	v1alpha1.MockServer
	MockResponse(ctx context.Context, request *interact.Request) (*interact.Response, error)
	Start(ctx context.Context, cancelFunc context.CancelFunc) error
}

// Manager is the implement of APIManager
type Manager struct {
	cfg *Config

	storage        pluginregistry.StoragePlugin
	pluginRegistry pluginregistry.Registry
	// does not support deletion
	// https://github.com/gorilla/mux/issues/82
	// readonly
	mux *mux.Router
	// map[uniqueKey]*v1alpha1.MockAPI
	// readonly
	apis map[string]*v1alpha1.MockAPI
	// used to protect the pointer of mux, apis
	lock sync.RWMutex

	v1alpha1.UnimplementedMockServer
	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	GRPCAddress string
	HTTPAddress string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		GRPCAddress: "0.0.0.0:30000",
		HTTPAddress: "0.0.0.0:30001",
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.GRPCAddress, prefix+"apiManager.grpcAddress", c.GRPCAddress, "gRPC service listener address")
	f.StringVar(&c.HTTPAddress, prefix+"apiManager.httpAddress", c.HTTPAddress, "http service listener address")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.HTTPAddress == "" && c.GRPCAddress == "" {
		return errors.New("[apiManager] grpcAddress and httpAddress cannot be empty at the same time")
	}
	return nil
}

// New is used to init service
func New(cfg *Config,
	pluginRegistry pluginregistry.Registry,
	logger logger.Logger, registerer prometheus.Registerer) (Provider, error) {
	service := &Manager{
		cfg:            cfg,
		registerer:     registerer,
		pluginRegistry: pluginRegistry,
		mux:            mux.NewRouter(),
		apis:           map[string]*v1alpha1.MockAPI{},
		Logger:         logger.NewLogger("apiManager"),
	}
	return service, nil
}

// Start is used to start the service
func (s *Manager) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	if err := s.setupStorage(); err != nil {
		return err
	}
	if err := s.loadAPIs(ctx); err != nil {
		return err
	}
	if err := s.setupGRPCServer(ctx, cancelFunc); err != nil {
		return err
	}
	if err := s.setupHTTPServer(ctx, cancelFunc); err != nil {
		return err
	}
	s.setupAnnouncementReceiver(ctx, cancelFunc)
	return nil
}

// SaveMockAPI is used to create or update MockAPI
func (s *Manager) SaveMockAPI(ctx context.Context, request *v1alpha1.SaveMockAPIRequest) (*v1alpha1.SaveMockAPIResponse, error) {
	api := request.GetData()
	if api == nil {
		return nil, errors.New("api is nil")
	}
	var encoder jsonpb.Marshaler
	data, err := encoder.MarshalToString(api)
	if err != nil {
		return nil, err
	}
	if err := s.storage.Set(ctx, api.GetUniqueKey(), data); err != nil {
		return nil, err
	}
	return &v1alpha1.SaveMockAPIResponse{}, nil
}

// DeleteMockAPI is used to delete MockAPI
func (s *Manager) DeleteMockAPI(ctx context.Context, request *v1alpha1.DeleteMockAPIRequest) (*v1alpha1.DeleteMockAPIResponse, error) {
	uniqueKey := request.GetUniqueKey()
	if err := s.storage.Delete(ctx, uniqueKey); err != nil {
		return nil, err
	}
	return &v1alpha1.DeleteMockAPIResponse{}, nil
}

// ListMockAPI is used to list MockAPIs
func (s *Manager) ListMockAPI(ctx context.Context, request *v1alpha1.ListMockAPIRequest) (*v1alpha1.ListMockAPIResponse, error) {
	s.lock.RLock()
	apis := s.apis
	s.lock.RUnlock()

	var total uint64
	var uniqueKeys []string
	keywords := request.GetKeywords()

	for _, mockAPI := range apis {
		total++
		if keywords != "" && !strings.Contains(mockAPI.GetUniqueKey(), keywords) {
			continue
		}
		uniqueKeys = append(uniqueKeys, mockAPI.GetUniqueKey())
	}

	sort.Strings(uniqueKeys)

	pagination := util.GetPagination(request.GetPagination())
	if err := util.PaginateSlice(pagination, &uniqueKeys); err != nil {
		return nil, err
	}

	data := make([]*v1alpha1.MockAPI, 0, len(uniqueKeys))
	for _, key := range uniqueKeys {
		mockAPI, ok := apis[key]
		if ok {
			data = append(data, mockAPI)
		}
	}
	return &v1alpha1.ListMockAPIResponse{
		Data: data,
	}, nil
}

// MatchAPI is used to match MockAPI
func (s *Manager) MatchAPI(host, path, method string) (*v1alpha1.MockAPI, bool) {
	s.lock.RLock()
	m := s.mux
	apis := s.apis
	s.lock.RUnlock()

	var match mux.RouteMatch
	matched := m.Match(&http.Request{
		Method: method,
		URL:    &url.URL{Path: path},
		Host:   host,
	}, &match)
	if !matched {
		return nil, false
	}

	api := apis[match.Route.GetName()]
	if api != nil {
		return api, true
	}
	return nil, false
}

// MockResponse is used to mock response
func (s *Manager) MockResponse(ctx context.Context, request *interact.Request) (*interact.Response, error) {
	api, ok := s.MatchAPI(request.Host, request.Path, request.Method)
	if !ok {
		return nil, fmt.Errorf("unable to find mock config of %s", request.Path)
	}
	mockCase, err := s.getMatchedCase(ctx, request, api)
	if err != nil {
		return nil, err
	}
	response := interact.NewDefaultResponse(request)
	for _, plugin := range s.pluginRegistry.MockPlugins() {
		abort, err := plugin.MockResponse(ctx, mockCase.GetResponse(), request, response)
		if err != nil {
			return nil, newPluginError(codes.Internal, plugin.Name(), err)
		}
		if abort {
			return response, nil
		}
	}
	return response, nil
}

func (s *Manager) setupStorage() error {
	storagePlugin := s.pluginRegistry.StoragePlugin()
	if storagePlugin != nil {
		s.storage = storagePlugin
		return nil
	}
	storage, err := memory.New(memory.NewConfig(), s.NewLogger(".memory"), s.registerer)
	if err != nil {
		return err
	}
	s.storage = storage
	return nil
}

func (s *Manager) setupHTTPServer(ctx context.Context, cancelFunc func()) error {
	addr := s.cfg.HTTPAddress
	if addr == "" {
		return nil
	}
	s.LogInfo(nil, "starting api manager on http address: %s", addr)
	serverMux := runtime.NewServeMux()
	err := v1alpha1.RegisterMockHandlerFromEndpoint(context.TODO(), serverMux, s.cfg.GRPCAddress, []grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:    s.cfg.HTTPAddress,
		Handler: serverMux,
	}
	util.StartServiceAsync(ctx, cancelFunc, s.Logger.NewLogger("http"), func() error {
		return server.ListenAndServe()
	}, func() error {
		return server.Shutdown(context.TODO())
	})
	return nil
}

func (s *Manager) setupGRPCServer(ctx context.Context, cancelFunc func()) error {
	addr := s.cfg.GRPCAddress
	if addr == "" {
		return nil
	}
	s.LogInfo(nil, "starting api manager on gRPC address: %s", addr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(util.GRPCLoggingMiddleware(s.Logger)))
	v1alpha1.RegisterMockServer(server, s)
	util.StartServiceAsync(ctx, cancelFunc, s.Logger.NewLogger("gRPC"), func() error {
		return server.Serve(listener)
	}, func() error {
		server.GracefulStop()
		return nil
	})
	return nil
}

func (s *Manager) setupAnnouncementReceiver(ctx context.Context, cancelFunc func()) {
	util.StartServiceAsync(ctx, cancelFunc, s.Logger, func() error {
		for {
			select {
			case _, ok := <-s.storage.GetAnnouncement():
				if !ok {
					s.LogWarn(nil, "storage announcement closed")
					return nil
				}
				s.LogInfo(nil, "storage announcement received")
				if err := s.loadAPIs(ctx); err != nil {
					s.LogError(nil, "failed to load apis: %s", err)
				}
			case <-ctx.Done():
				s.LogWarn(nil, "apiManager stops watching announcements")
				return nil
			}
		}
	}, func() error {
		return nil
	})
}

func (s *Manager) loadAPIs(ctx context.Context) error {
	pairs, err := s.storage.List(ctx)
	if err != nil {
		return err
	}
	apis := map[string]*v1alpha1.MockAPI{}
	s.LogInfo(nil, "load apis from storage, total %d", len(pairs))
	for key, val := range pairs {
		var api v1alpha1.MockAPI
		if err := jsonpb.UnmarshalString(val, &api); err != nil {
			return fmt.Errorf("failed to load(%s): %s", key, err)
		}
		apis[key] = &api
		s.LogInfo(map[string]interface{}{
			"uniqueKey": api.GetUniqueKey(),
			"path":      api.GetPath(),
		}, "apis is loaded")
	}
	s.lock.Lock()
	s.apis = apis
	s.mux = buildMux(apis, s.Logger)
	s.lock.Unlock()
	return nil
}

func (s *Manager) getMatchedCase(ctx context.Context, request *interact.Request, api *v1alpha1.MockAPI) (*v1alpha1.MockAPI_Case, error) {
	for _, mockCase := range api.Cases {
		for _, plugin := range s.pluginRegistry.MatchPlugins() {
			condition := mockCase.GetCondition()
			if condition == nil {
				return mockCase, nil
			}
			matched, err := plugin.Match(ctx, request, condition)
			if err != nil {
				return nil, newPluginError(codes.Internal, plugin.Name(), err)
			}
			if matched {
				return mockCase, nil
			}
		}
	}
	return nil, status.Error(codes.NotFound, "no case matched")
}

func buildMux(apis map[string]*v1alpha1.MockAPI, log logger.Logger) *mux.Router {
	router := mux.NewRouter()
	for _, mockAPI := range apis {
		if err := addAPI(router, mockAPI); err != nil {
			log.LogWarn(map[string]interface{}{
				"uniqueKey": mockAPI.GetUniqueKey(),
			}, "failed to add api when buildMux: %s", err)
		}
	}
	return router
}

func addAPI(router *mux.Router, api *v1alpha1.MockAPI) error {
	if api.GetUniqueKey() == "" {
		return errors.New("unique key is required")
	}
	if api.Path == "" {
		return errors.New("path is required")
	}
	route := router.Path(api.Path)
	if api.Host != "" {
		route = route.Host(api.Host)
	}
	if api.Method != "" {
		route = route.Methods(api.Method)
	}
	route.Name(api.UniqueKey)
	return nil
}

func newPluginError(code codes.Code, name string, err error) error {
	return status.Error(code, fmt.Sprintf("plugin(%s): %s", name, err))
}
