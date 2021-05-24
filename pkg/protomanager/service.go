package protomanager

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/storyicon/powermock/pkg/util/logger"
)

// Provider is used to read, parse and manage Proto files
type Provider interface {
	// GetMethod is used to get descriptor of specified grpc path
	GetMethod(name string) (*desc.MethodDescriptor, bool)
}

// Manager is the implement of Provider
type Manager struct {
	cfg *Config

	// map[name]*desc.MethodDescriptor
	methods sync.Map

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	ProtoImportPaths []string
	ProtoDir         string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	var protoImportPaths []string
	if goPath := os.Getenv("GOPATH"); goPath != "" {
		protoImportPaths = append(protoImportPaths, goPath, filepath.Join(goPath, "bin"))
	}
	return &Config{
		ProtoImportPaths: protoImportPaths,
		ProtoDir:         "./",
	}
}

// RegisterFlags is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringArrayVar(&c.ProtoImportPaths, prefix+"protoManager.protoImportPaths", c.ProtoImportPaths, "proto import paths")
	f.StringVar(&c.ProtoDir, prefix+"protoManager.protoDir", c.ProtoDir, "proto dir to load")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.ProtoDir == "" {
		return errors.New("protoDir is required")
	}
	return nil
}

// New is used to init service
func New(cfg *Config, logger logger.Logger, registerer prometheus.Registerer) (*Manager, error) {
	service := &Manager{
		cfg:        cfg,
		registerer: registerer,
		Logger:     logger.NewLogger("protoManager"),
	}
	if err := service.loadProto(); err != nil {
		return nil, err
	}
	return service, nil
}

// GetMethod is used to get descriptor of specified grpc path
func (s *Manager) GetMethod(name string) (*desc.MethodDescriptor, bool) {
	val, ok := s.methods.Load(name)
	if !ok {
		return nil, false
	}
	return val.(*desc.MethodDescriptor), true
}

func (s *Manager) addMethod(name string, method *desc.MethodDescriptor) error {
	_, loaded := s.methods.LoadOrStore(name, method)
	if loaded {
		return fmt.Errorf("method already exists: %s", name)
	}
	return nil
}

func (s *Manager) loadProto() error {
	protoDir := s.cfg.ProtoDir
	importPaths := append([]string{protoDir}, s.cfg.ProtoImportPaths...)
	s.LogInfo(nil, "starting load proto from: %s", protoDir)
	return filepath.Walk(protoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			s.LogWarn(nil, "load proto error: %s", err)
			// * skip fs error
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".proto" {
			return nil
		}
		relPath, err := filepath.Rel(protoDir, path)
		if err != nil {
			return err
		}
		parser := protoparse.Parser{
			ImportPaths:           importPaths,
			InferImportPaths:      len(importPaths) == 0,
			IncludeSourceCodeInfo: true,
		}
		fds, err := parser.ParseFiles(relPath)
		if err != nil {
			s.LogError(nil, "failed to parse file: %s", err)
			return nil
		}
		for _, fd := range fds {
			for _, service := range fd.GetServices() {
				for _, method := range service.GetMethods() {
					name := GetPathByFullyQualifiedName(method.GetFullyQualifiedName())
					s.LogInfo(map[string]interface{}{
						"name": name,
					}, "api loaded")
					if err := s.addMethod(name, method); err != nil {
						s.LogWarn(map[string]interface{}{
							"name":  name,
							"error": err,
						}, "failed to load method")
					}
				}
			}
		}
		return nil
	})
}

// GetPathByFullyQualifiedName is used to get the grpc path of specified fully qualified name
func GetPathByFullyQualifiedName(name string) string {
	raw := []byte(name)
	if i := bytes.LastIndexByte(raw, '.'); i > 0 {
		raw[i] = '/'
	}
	return string(append([]byte{'/'}, raw...))
}
