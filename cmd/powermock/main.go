package main

import (
	"fmt"

	"github.com/spf13/cobra"

	cmdsload "github.com/bilibili-base/powermock/cmd/powermock/cmds/load"
	cmdsserve "github.com/bilibili-base/powermock/cmd/powermock/cmds/serve"
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

var cmdRoot = &cobra.Command{
	Use:     "powermock",
	Short:   "powermock",
	Version: fmt.Sprintf("%s, branch: %s, revision: %s, buildDate: %s", Version, Branch, Revision, BuildDate),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(asciiImage)
		_ = cmd.Help()
	},
}

func main() {
	cmdRoot.AddCommand(cmdsserve.CmdServe, cmdsload.CmdLoad)
	_ = cmdRoot.Execute()
}
