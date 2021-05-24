package main

import (
	"context"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	"github.com/storyicon/powermock/cmd/powermock/internal"
	"github.com/storyicon/powermock/pkg/util"
	"github.com/storyicon/powermock/pkg/util/logger"
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
Powered by: storyicon

`

var config = internal.NewConfig()
var cmd = cobra.Command{
	Use:     "",
	Version: fmt.Sprintf(`Version: %s, Branch: %s, Revision: %s, BuildDate: %s`, Version, Branch, Revision, BuildDate),
	Short:   "",
	Run: func(cmd *cobra.Command, args []string) {
		log, err := logger.New(config.Log, "main", prometheus.DefaultRegisterer)
		if err != nil {
			panic(err)
		}
		fmt.Println(asciiImage)

		ctx, cancel := context.WithCancel(context.Background())
		stop := util.RegisterExitHandlers(log, cancel)
		defer cancel()

		util.DumpYaml(config)

		if err := config.Validate(); err != nil {
			log.LogFatal(nil, "failed to validate config: %s", err)
		}

		log.LogInfo(map[string]interface{}{
			"version":   Version,
			"branch":    Branch,
			"revision":  Revision,
			"buildDate": BuildDate,
		}, "Welcome to PowerMock")
		if err := internal.Startup(ctx, cancel, config, log, prometheus.DefaultRegisterer); err != nil {
			log.LogFatal(nil, "oops, an error has occurred: %s", err)
		}
		<-stop
		log.LogInfo(nil, "Goodbye")
	},
}

func main() {
	configFile := util.ParseConfigFileParameter(os.Args[1:])
	if configFile != "" {
		fmt.Printf("start to load config file: %s \r\n", configFile)
		if err := util.LoadConfig(configFile, &config); err != nil {
			fmt.Printf("error loading config from %s: %v\n", configFile, err)
			os.Exit(1)
		}
	}
	flagSet := cmd.PersistentFlags()
	util.IgnoredFlag(flagSet, "config.file", "")
	config.RegisterFlagsWithPrefix("", flagSet)
	_ = cmd.Execute()
}
