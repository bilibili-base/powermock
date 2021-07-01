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

package main

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	cmdsload "github.com/bilibili-base/powermock/cmd/powermock/subcommands/load"
	cmdsserve "github.com/bilibili-base/powermock/cmd/powermock/subcommands/serve"
	bootstrap "github.com/bilibili-base/powermock/pkg/bootstraps/v8"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

// Version is set via build flag -ldflags -X main.Version
var (
	Version   string
	Branch    string
	Revision  string
	BuildDate string
)

var asciiImage = `
------------------------------------
    ____                          __  ___           __  
   / __ \____ _      _____  _____/  |/  /___  _____/ /__
  / /_/ / __ \ | /| / / _ \/ ___/ /|_/ / __ \/ ___/ //_/
 / ____/ /_/ / |/ |/ /  __/ /  / /  / / /_/ / /__/ ,<   
/_/    \____/|__/|__/\___/_/  /_/  /_/\____/\___/_/|_|  
                                                        
------------------------------------
Powered by: bilibili
compiled with V8
`

var cmdRoot = &cobra.Command{
	Use:     "powermock",
	Short:   "powermock",
	Version: fmt.Sprintf("%s, branch: %s, revision: %s, buildDate: %s", Version, Branch, Revision, BuildDate),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(asciiImage)
		_ = cmd.Help()
	},
}

var config = bootstrap.NewConfig()

func Setup(ctx context.Context) {
	log, err := logger.New(config.Log, "main", prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	stop := util.RegisterExitHandlers(log, cancel)
	defer cancel()

	util.DumpYaml(config)

	if err := config.Validate(); err != nil {
		log.LogFatal(nil, "failed to validate config: %s", err)
	}
	if err := bootstrap.Startup(ctx, cancel, config, log, prometheus.DefaultRegisterer); err != nil {
		log.LogFatal(nil, "oops, an error has occurred: %s", err)
	}
	<-stop
	log.LogInfo(nil, "Goodbye")
}

func main() {
	cmdRoot.AddCommand(
		cmdsserve.CommandServe(Setup, config),
		cmdsload.CommandLoad(),
	)
	_ = cmdRoot.Execute()
}
