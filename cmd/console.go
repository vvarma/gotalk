package cmd

import (
	"github.com/ipfs/go-log/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/vvarma/gotalk/console"
)

func init() {
	rootCmd.AddCommand(consoleCmd)
}

var (
	consoleCmd = &cobra.Command{
		Use: "console",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetupLogging(log.Config{
				Stderr: false,
				Stdout: false,
			})
			log.SetAllLoggers(log.LevelWarn)
			log.SetLogLevel("commands", "debug")
			log.SetLogLevel("gotalk", "debug")
			log.SetLogLevel("paraU", "debug")
			log.SetLogLevel("subcmd", "debug")
			log.SetLogLevel("chat", "debug")
			log.SetLogLevel("client", "debug")
			log.SetLogLevel("control", "debug")
			log.SetLogLevel("dost", "debug")
			_ = log.SetLogLevel("console", "debug")
			app := tview.NewApplication().EnableMouse(true)
			view := console.New()
			app.SetRoot(view.ParentFlex, true)
			return app.Run()
		},
	}
)
