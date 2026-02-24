package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start [profile_id]",
	Short: "Start profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		err := sprofile.Start(args[0], mode, password, passwordPrompt)
		cobra.CheckErr(err)
	},
}
