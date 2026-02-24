package cmd

import (
	"github.com/pritunl/pritunl-client/cli/watch"
	"github.com/spf13/cobra"
)

var WatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch profiles",
	Run: func(cmd *cobra.Command, args []string) {
		err := watch.Init()
		cobra.CheckErr(err)
	},
}
