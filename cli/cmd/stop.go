package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop [profile_id]",
	Short: "Stop profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		err := sprofile.Stop(args[0])
		cobra.CheckErr(err)
	},
}
