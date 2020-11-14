package cmd

import (
	"github.com/spf13/cobra"
	"github.com/vvarma/gotalk/pkg/metrics"
)

var (
	metricsEnabled bool
	rootCmd        = &cobra.Command{
		Use: "gotalk",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if metricsEnabled {
				exp, err := metrics.NewExporter()
				if err != nil {
					logger.Fatal(err)
				}
				metrics.ExporterInstance = exp
				exp.Start()
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize()
	rootCmd.PersistentFlags().BoolVarP(&metricsEnabled, "metrics", "m", false, "Enable metrics")
}
