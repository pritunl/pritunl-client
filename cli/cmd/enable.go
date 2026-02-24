package cmd

import (
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var EnableCmd = &cobra.Command{
	Use:   "enable [profile_id]",
	Short: "Enable autostart for profile",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile ID")
		}

		err := sprofile.SetState(args[0], true)
		cobra.CheckErr(err)
	},
}
