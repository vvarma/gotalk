package cmd

import "github.com/spf13/cobra"

var(
	username string
)
func userNameFlag(cmd*cobra.Command){
	cmd.Flags().StringVarP(&userName, "user", "u", "", "username")
	_ = cmd.MarkFlagRequired("user")

}
