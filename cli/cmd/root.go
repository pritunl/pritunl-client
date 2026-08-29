package cmd

import (
	"github.com/pritunl/pritunl-client/cli/iface"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "pritunl-client",
	Short: "Pritunl Client Command Line Tool",
	Long: "Pritunl Client Command Line Tool\n\n" +
		"Run without a command to open the interactive terminal interface.",
	SilenceErrors: true,
	SilenceUsage:  true,
	Run: func(cmd *cobra.Command, args []string) {
		err := iface.Iface()
		cobra.CheckErr(err)
	},
}

func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(AddCmd)
	RootCmd.AddCommand(RemoveCmd)
	RootCmd.AddCommand(EnableCmd)
	RootCmd.AddCommand(DisableCmd)
	RootCmd.AddCommand(LogsCmd)
	RootCmd.AddCommand(ListCmd)
	RootCmd.AddCommand(StartCmd)
	RootCmd.AddCommand(StopCmd)
}
