package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var DisableCmd = &cobra.Command{
	Use:   "disable [profile_id]",
	Short: "Disable autostart for profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		err := sprofile.SetState(args[0], false)
		cobra.CheckErr(err)
	},
}
