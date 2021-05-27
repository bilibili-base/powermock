// Copyright 2020 MOSS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

// Logger defines the basic log library implementation
type Logger interface {
	// LogDebug print a message with debug level.
	LogDebug(fields map[string]interface{}, format string, args ...interface{})
	// LogInfo print a message with info level.
	LogInfo(fields map[string]interface{}, format string, args ...interface{})
	// LogWarn print a message with warn level.
	LogWarn(fields map[string]interface{}, format string, args ...interface{})
	// LogError print a message with error level.
	LogError(fields map[string]interface{}, format string, args ...interface{})
	// LogFatal print a message with fatal level.
	LogFatal(fields map[string]interface{}, format string, args ...interface{})
	// NewLogger is used to derive a new child Logger
	NewLogger(component string) Logger
	// SetLogLevel is used to set log level
	SetLogLevel(verbosity string)
}

// BasicLogger simply implements Logger
type BasicLogger struct {
	cfg *Config

	component  string
	logger     zerolog.Logger
	registerer prometheus.Registerer
}

// Config defines the config structure
type Config struct {
	Pretty bool
	Level  string
}

// NewConfig is used to init config with default values
func NewConfig() *Config {
	return &Config{
		Pretty: true,
		Level:  "debug",
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.Level, prefix+"log.level", c.Level, "log level(debug, info, warn, error, fatal)")
	f.BoolVar(&c.Pretty, prefix+"log.pretty", c.Pretty, "log in a pretty format")
}

// NewDefault is used to initialize a simple Logger
func NewDefault(component string) Logger {
	logger, err := New(NewConfig(), component, prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}
	return logger
}

// New is used to init service
func New(cfg *Config, component string, registerer prometheus.Registerer) (Logger, error) {
	if cfg == nil {
		cfg = NewConfig()
	}
	service := &BasicLogger{
		cfg:        cfg,
		component:  component,
		registerer: registerer,
	}
	service.setup()
	return service, nil
}

func (b *BasicLogger) setup() {
	b.logger = log.With().Str("component", b.component).Logger().Hook(CallerHook{})
	if b.cfg != nil {
		if b.cfg.Pretty {
			b.logger = b.logger.Output(zerolog.ConsoleWriter{
				Out: os.Stdout,
			})
		}
		b.SetLogLevel(b.cfg.Level)
	}
}

// LogDebug print a message with debug level.
func (b *BasicLogger) LogDebug(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Debug().Fields(fields).Msgf(format, args...)
}

// LogInfo print a message with info level.
func (b *BasicLogger) LogInfo(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Info().Fields(fields).Msgf(format, args...)
}

// LogWarn print a message with warn level.
func (b *BasicLogger) LogWarn(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Warn().Fields(fields).Msgf(format, args...)
}

// LogError print a message with error level.
func (b *BasicLogger) LogError(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Error().Fields(fields).Msgf(format, args...)
}

// LogFatal print a message with fatal level.
func (b *BasicLogger) LogFatal(fields map[string]interface{}, format string, args ...interface{}) {
	b.logger.Fatal().Fields(fields).Msgf(format, args...)
}

// NewLogger is used to derive a new child Logger
func (b *BasicLogger) NewLogger(component string) Logger {
	name := strings.Join([]string{b.component, component}, ".")
	logger, err := New(b.cfg, name, b.registerer)
	if err != nil {
		b.LogWarn(map[string]interface{}{
			"name": name,
		}, "failed to extend logger: %s", err)
		return b
	}
	return logger
}

// SetLogLevel is used to set log level
func (b *BasicLogger) SetLogLevel(verbosity string) {
	switch verbosity {
	case "debug":
		b.logger.Level(zerolog.DebugLevel)
	case "info":
		b.logger.Level(zerolog.InfoLevel)
	case "warn":
		b.logger.Level(zerolog.WarnLevel)
	case "error":
		b.logger.Level(zerolog.ErrorLevel)
	case "fatal":
		b.logger.Level(zerolog.FatalLevel)
	}
}

// CallerHook implements zerolog.Hook interface.
type CallerHook struct{}

// Run adds additional context
func (h CallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if _, file, line, ok := runtime.Caller(4); ok {
		e.Str("file", fmt.Sprintf("%s:%d", path.Base(file), line))
	}
}
