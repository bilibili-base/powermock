package http

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/apimanager"
	"github.com/bilibili-base/powermock/pkg/interact"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Provider defines the mock server interface
type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc) error
}

// MockServer is the implement of http mock server
type MockServer struct {
	cfg        *Config
	apiManager apimanager.Provider
	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	Enable  bool
	Address string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Enable:  true,
		Address: "0.0.0.0:30003",
	}
}

// IsEnabled is used to return whether the current component is enabled
// This attribute is required in pluggable components
func (c *Config) IsEnabled() bool {
	return c.Enable
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.BoolVar(&c.Enable, prefix+"httpMockServer.enable", c.Enable, "define whether the component is enabled")
	f.StringVar(&c.Address, prefix+"httpMockServer.address", c.Address, "address to listen")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("the address of mockserver is required")
	}
	return nil
}

// New is used to init service
func New(cfg *Config,
	apiManager apimanager.Provider,
	logger logger.Logger, registerer prometheus.Registerer) (Provider, error) {
	service := &MockServer{
		cfg:        cfg,
		registerer: registerer,
		apiManager: apiManager,
		Logger:     logger.NewLogger("httpMockServer"),
	}
	return service, nil
}

// ServeHTTP is used to implement the interface of http.Handler
func (s *MockServer) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	s.LogInfo(map[string]interface{}{
		"path":    request.URL.String(),
		"host":    request.Host,
		"headers": request.Header,
		"method":  request.Method,
	}, "request received")
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	resp, err := s.apiManager.MockResponse(request.Context(), &interact.Request{
		Protocol: interact.ProtocolHTTP,
		Method:   request.Method,
		Host:     request.Host,
		Path:     request.URL.Path,
		Header:   getHeadersFromHttpHeaders(request.Header),
		Body:     interact.NewBytesMessage(body),
	})
	if err != nil {
		sendError(w, util.GetHTTPCodeFromError(err), err)
		return
	}
	if code := resp.Code; code >= 100 && code <= 999 {
		w.WriteHeader(int(resp.Code))
	}
	for key, val := range resp.Header {
		w.Header().Set(key, val)
	}
	w.Write(resp.Body.Bytes())
}

// Start is used to start the service
func (s *MockServer) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	s.LogInfo(nil, "starting http mock server on: %s", s.cfg.Address)
	listener, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		return err
	}
	util.StartServiceAsync(ctx, cancelFunc, s.Logger, func() error {
		return http.Serve(listener, s)
	}, func() error {
		return listener.Close()
	})
	return nil
}

func sendError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

// getHeadersFromHttpHeaders is used to get map[string]string from http.Header
func getHeadersFromHttpHeaders(input http.Header) map[string]string {
	headers := map[string]string{}
	for key, values := range input {
		key = strings.ToLower(key)
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
