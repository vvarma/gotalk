package cmd

import (
	"github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/vvarma/gotalk/pkg/gotalk"
)

var logger = log.Logger("commands")

func init() {
	rootCmd.AddCommand(chatCommand)
	chatCommand.Flags().StringVarP(&randevousString, "randevous1", "r", "default randevous", "Randevous string")
	chatCommand.Flags().StringVarP(&userName, "user", "u", "user1", "username")
}

var (
	randevousString string
	userName        string
	chatCommand     = &cobra.Command{
		Use: "chat",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := gotalk.NewChat(userName, randevousString)
			if err != nil {
				logger.Fatal("Error starting chat command", err)
			}
			go c.Input()
			select {}
		},
	}
)
