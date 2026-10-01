package cmd

import (
	"github.com/pritunl/pritunl-client/cli/constants"
	"github.com/pritunl/pritunl-client/cli/sprofile"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:   "add [profile_uri|tar_path|ovpn_path]",
	Short: "Add profile",
	Long: "Add profile\n\n" +
		"Profiles are added as system profiles stored by the service " +
		"unless --user is set. Flatpak installs only support user profiles. " +
		"Autostart is disabled on new system profiles unless --enable is " +
		"set or the server enforces it.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cobra.CheckErr("cmd: Missing profile URI or path")
		}

		system := !userProfile && !constants.Flatpak

		if enableProfile && !system {
			if constants.Flatpak {
				cobra.CheckErr("cmd: Autostart requires a system profile, " +
					"Flatpak installs only support user profiles")
			}
			cobra.CheckErr("cmd: Autostart requires a system profile, " +
				"--enable cannot be used with --user")
		}

		err := sprofile.ImportPath(args[0], system, enableProfile)
		cobra.CheckErr(err)
	},
}
