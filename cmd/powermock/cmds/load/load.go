package load

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"

	yamltool "github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"

	"github.com/storyicon/powermock/apis/v1alpha1"
	"github.com/storyicon/powermock/pkg/util"
	"github.com/storyicon/powermock/pkg/util/logger"
)

type Config struct {
	Address string
	File    string
}

func NewConfig() *Config {
	return &Config{
		Address: "127.0.0.1:30000",
		File:    "",
	}
}

// RegisterFlagsWithPrefix is used to register flags
func (c *Config) RegisterFlagsWithPrefix(prefix string, f *pflag.FlagSet) {
	f.StringVar(&c.Address, "address", c.Address, "the mock server listen address to call")
	f.StringVar(&c.File, "file", c.File, "file path, yaml format is supported")
}

// Validate is used to validate config and returns error on failure
func (c *Config) Validate() error {
	if c.Address == "" {
		return errors.New("address is required")
	}
	if c.File == "" {
		return errors.New("file is required")
	}
	return nil
}

var config = NewConfig()
var CmdLoad = &cobra.Command{
	Use:   "load",
	Short: "load mock api from file",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewDefault("main")
		if err := config.Validate(); err != nil {
			log.LogFatal(nil, "failed to validate config: %s", err)
		}

		log.LogInfo(nil, "start to load file")
		apis, err := loadMockAPIs(log, config.File)
		if err != nil {
			log.LogFatal(nil, "failed to load apis(%s): %s", config.File, err)
		}
		log.LogInfo(map[string]interface{}{
			"count": len(apis),
		}, "mock apis loaded from file")

		conn, err := grpc.Dial(config.Address, grpc.WithInsecure())
		if err != nil {
			log.LogWarn(nil, "please start the mock service through the `powermock serve` command first, "+
				"and make sure that the correct `address` is specified")
			log.LogFatal(nil, "failed to dial: %s", err)
		}
		client := v1alpha1.NewMockClient(conn)
		for _, api := range apis {
			log.LogInfo(map[string]interface{}{
				"uniqueKey": api.GetUniqueKey(),
				"method":    api.GetMethod(),
				"host":      api.GetHost(),
				"path":      api.GetPath(),
			}, "start to save api")
			_, err = client.SaveMockAPI(context.TODO(), &v1alpha1.SaveMockAPIRequest{
				Data: api,
			})
			if err != nil {
				log.LogFatal(nil, "failed to save mock api: %s", err)
			}
		}
		log.LogInfo(nil, "succeed!")
	},
}

func init() {
	config.RegisterFlagsWithPrefix("", CmdLoad.PersistentFlags())
}

func loadMockAPIs(log logger.Logger, file string) ([]*v1alpha1.MockAPI, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.LogFatal(nil, "failed to load file(%s): %s", config.File, err)
	}
	parts, err := util.SplitYAML(data)
	if err != nil {
		return nil, err
	}
	var apis []*v1alpha1.MockAPI
	for _, part := range parts {
		data, err := yamltool.YAMLToJSON(part)
		if err != nil {
			return nil, err
		}
		var api v1alpha1.MockAPI
		if err = jsonpb.Unmarshal(bytes.NewReader(data), &api); err != nil {
			return nil, err
		}
		apis = append(apis, &api)
	}
	return apis, nil
}
