package serve

import (
	"context"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	"github.com/bilibili-base/powermock/cmd/powermock/internal"
	"github.com/bilibili-base/powermock/pkg/util"
	"github.com/bilibili-base/powermock/pkg/util/logger"
)

var config = internal.NewConfig()

var CmdServe = &cobra.Command{
	Use:   "serve",
	Short: "start the mock server",
	Run: func(cmd *cobra.Command, args []string) {
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
		if err := internal.Startup(ctx, cancel, config, log, prometheus.DefaultRegisterer); err != nil {
			log.LogFatal(nil, "oops, an error has occurred: %s", err)
		}
		<-stop
		log.LogInfo(nil, "Goodbye")
	},
}

func init() {
	configFile := util.ParseConfigFileParameter(os.Args[1:])
	if configFile != "" {
		fmt.Printf("start to load config file: %s \r\n", configFile)
		if err := util.LoadConfig(configFile, &config); err != nil {
			fmt.Printf("error loading config from %s: %v\n", configFile, err)
			os.Exit(1)
		}
	}
	flagSet := CmdServe.PersistentFlags()
	util.IgnoredFlag(flagSet, "config.file", "")
	config.RegisterFlagsWithPrefix("", flagSet)
}
