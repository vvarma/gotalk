package cmd

import (
	"context"
	"github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"github.com/vvarma/gotalk/pkg/gotalk"
)

var logger = log.Logger("commands")

func init() {
	rootCmd.AddCommand(chatCommand)
	chatCommand.Flags().StringVarP(&randevousString, "randevous1", "r", "default randevous", "Randevous string")
	chatCommand.Flags().StringVarP(&listenAddress, "listen", "l", "", "listen address")
	chatCommand.Flags().StringVarP(&mode, "mode", "i", "chat", "load or chat")
	userNameFlag(chatCommand)
}

var (
	randevousString string
	userName        string
	listenAddress   string
	mode            string
	chatCommand     = &cobra.Command{
		Use: "chat",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			_, err := gotalk.Open(ctx, userName)
			//c, err := gotalk.NewChat(userName, randevousString, mode)
			if err != nil {
				logger.Fatal("Error starting chat command", err)
			}
			//go c.Input()
			select {}
		},
	}
)
