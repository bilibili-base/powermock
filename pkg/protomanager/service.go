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

package protomanager

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/bilibili-base/powermock/pkg/protomanager/synchronize"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Provider is used to read, parse and manage Proto files
type Provider interface {
	Start(ctx context.Context, cancelFunc context.CancelFunc) error
	// GetMethod is used to get descriptor of specified grpc path
	GetMethod(name string) (*desc.MethodDescriptor, bool)
}

// Manager is the implement of Provider
type Manager struct {
	cfg *Config

	// map[name]*desc.MethodDescriptor
	methods         *sync.Map
	methodsLock     sync.Mutex
	synchronization *synchronize.Service

	registerer prometheus.Registerer
	logger.Logger
}

// Config defines the config structure
type Config struct {
	ProtoImportPaths []string
	ProtoDir         string
	Synchronization  *synchronize.Config
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
		Synchronization:  synchronize.NewConfig(),
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringArrayVar(&c.ProtoImportPaths, prefix+"protoManager.protoImportPaths", c.ProtoImportPaths, "proto import paths")
	f.StringVar(&c.ProtoDir, prefix+"protoManager.protoDir", c.ProtoDir, "proto dir to load")
	c.Synchronization.RegisterFlagsWithPrefix(prefix+"protoManager.synchronization.", f)
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
		methods:    &sync.Map{},
		Logger:     logger.NewLogger("protoManager"),
	}
	if cfg.Synchronization.Enable {
		synchronization, err := synchronize.New(cfg.Synchronization, logger, registerer)
		if err != nil {
			return nil, err
		}
		service.synchronization = synchronization
		if err := service.synchronizeProto(context.Background(), true); err != nil {
			return nil, err
		}
	} else {
		if err := service.loadProto(); err != nil {
			return nil, err
		}
	}
	return service, nil
}

// GetMethod is used to get descriptor of specified grpc path
func (s *Manager) GetMethod(name string) (*desc.MethodDescriptor, bool) {
	s.methodsLock.Lock()
	method := s.methods
	s.methodsLock.Unlock()
	val, ok := method.Load(name)
	if !ok {
		return nil, false
	}
	return val.(*desc.MethodDescriptor), true
}

func (s *Manager) Start(ctx context.Context, cancelFunc context.CancelFunc) error {
	if err := s.startSynchronization(ctx, cancelFunc); err != nil {
		return err
	}
	return nil
}

func (s *Manager) startSynchronization(ctx context.Context, cancelFunc context.CancelFunc) error {
	if !s.cfg.Synchronization.Enable {
		return nil
	}
	util.StartServiceAsync(ctx, cancelFunc, s.Logger, func() error {
		ticker := time.NewTicker(time.Minute * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.LogInfo(nil, "start to synchronize proto")
				if err := s.synchronizeProto(ctx, false); err != nil {
					s.LogError(map[string]interface{}{
						"err": err,
					}, "failed to synchronize proto")
				}
				s.LogInfo(nil, "synchronize proto finished")
			case <-ctx.Done():
				return nil
			}
		}
	}, func() error {
		return nil
	})
	return nil
}

func (s *Manager) synchronizeProto(ctx context.Context, force bool) error {
	var shouldReload bool
	_ = s.synchronization.Synchronize(ctx, func(repository string, updated bool, err error) error {
		if err != nil {
			s.LogWarn(map[string]interface{}{
				"error":      err,
				"repository": repository,
			}, "failed to synchronize repository")
			return nil
		}
		s.LogInfo(map[string]interface{}{
			"repository": repository,
			"updated":    updated,
		}, "repository synchronized")
		if updated {
			shouldReload = true
		}
		return nil
	})
	if shouldReload || force {
		if err := s.loadProto(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Manager) loadProto() error {
	var methods sync.Map
	protoDir := s.cfg.ProtoDir
	importPaths := append([]string{protoDir}, s.cfg.ProtoImportPaths...)
	s.LogInfo(nil, "starting load proto from: %s", protoDir)

	var count int
	err := filepath.Walk(protoDir, func(path string, info fs.FileInfo, err error) error {
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
					_, loaded := methods.LoadOrStore(name, method)
					if loaded {
						s.LogWarn(map[string]interface{}{
							"name":  name,
							"error": "method already exists",
						}, "failed to load method")
						continue
					}
					count++
				}
			}
		}
		return nil
	})

	s.LogInfo(map[string]interface{}{
		"total":    count,
		"protoDir": protoDir,
	}, "methods loaded")

	s.methodsLock.Lock()
	s.methods = &methods
	s.methodsLock.Unlock()

	if err != nil {
		return err
	}
	return nil
}

// GetPathByFullyQualifiedName is used to get the grpc path of specified fully qualified name
func GetPathByFullyQualifiedName(name string) string {
	raw := []byte(name)
	if i := bytes.LastIndexByte(raw, '.'); i > 0 {
		raw[i] = '/'
	}
	return string(append([]byte{'/'}, raw...))
}
