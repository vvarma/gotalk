package cmd

import (
	"bufio"
	"context"
	"github.com/spf13/cobra"
	"github.com/vvarma/gotalk/cmd/subcmd"
	"github.com/vvarma/gotalk/pkg/paraU"
	"github.com/vvarma/gotalk/pkg/paraU/client"
	"os"
	"strings"
)

func init() {
	userNameFlag(parUCmd)
	rootCmd.AddCommand(parUCmd)
}

var (
	parUCmd = &cobra.Command{
		Use: "paraU",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			opts := client.Options{Username: userName}
			client, err := paraU.New(ctx, opts)
			if err != nil {
				logger.Error("Error in execution", err)
			}
			ctx = paraU.SetInContext(ctx, client)
			for {
				r := bufio.NewReader(os.Stdin)
				line, err := r.ReadString('\n')
				if strings.HasSuffix(line, "\n") {
					line = strings.Split(line, "\n")[0]
				}
				if err != nil {
					logger.Error("Error in reading cmd ", err)
				}
				err = subcmd.Execute(ctx, line)
				if err != nil {
					logger.Error("Error executing cmd ", err)
				}
			}
		},
	}
)
