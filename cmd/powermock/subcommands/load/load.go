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

package load

import (
	"bytes"
	"context"
	"io/ioutil"

	yamltool "github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/bilibili-base/powermock/apis/v1alpha1"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

var (
	address = "127.0.0.1:30000"
)

var log = logger.NewDefault("commandline")

func CommandLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load mock api from file",
		Run: func(cmd *cobra.Command, args []string) {
			fileLocation := args[0]
			apis, err := loadMockAPIs(log, fileLocation)
			if err != nil {
				log.LogFatal(nil, "failed to load apis(%s): %s", fileLocation, err)
			}
			log.LogInfo(map[string]interface{}{
				"count": len(apis),
			}, "mock apis loaded from file")

			conn, err := grpc.Dial(address, grpc.WithInsecure())
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
	flag := cmd.PersistentFlags()
	flag.StringVar(&address, "address", address, "the gRPC address of mock server")
	return cmd
}

func loadMockAPIs(log logger.Logger, file string) ([]*v1alpha1.MockAPI, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.LogFatal(nil, "failed to load file(%s): %s", file, err)
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
