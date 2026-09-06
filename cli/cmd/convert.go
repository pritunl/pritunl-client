package cmd

import (
	"strings"

	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var ConvertCmd = &cobra.Command{
	Use:   "convert [profile_id] [system|user]",
	Short: "Convert profile between system and user profile",
	Long: "Convert profile between system and user profile\n\n" +
		"System profiles are stored by the service and can autostart, " +
		"user profiles are stored in the user data directory.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			cobra.CheckErr("cmd: Missing profile ID and profile type " +
				"(system or user)")
		}

		var system bool
		switch strings.ToLower(args[1]) {
		case "system":
			system = true
		case "user":
			system = false
		default:
			cobra.CheckErr("cmd: Invalid profile type, must be system or user")
		}

		err := sprofile.Convert(args[0], system)
		cobra.CheckErr(err)
	},
}
