package cmd

var (
	mode           string
	username       string
	password       string
	passwordPrompt bool
	userProfile    bool
	jsonFormat     bool
	jsonFormated   bool
)

func init() {
	StartCmd.Flags().StringVarP(
		&mode,
		"mode",
		"m",
		"",
		"VPN mode (ovpn, wg)",
	)
	StartCmd.Flags().StringVarP(
		&username,
		"username",
		"u",
		"",
		"VPN username for user profiles that require a username",
	)
	StartCmd.Flags().StringVarP(
		&password,
		"password",
		"p",
		"",
		"VPN password",
	)
	StartCmd.Flags().BoolVarP(
		&passwordPrompt,
		"password-read",
		"r",
		false,
		"Prompt for VPN password",
	)

	AddCmd.Flags().BoolVarP(
		&userProfile,
		"user",
		"u",
		false,
		"Add as user profile stored in the user data directory",
	)

	ListCmd.Flags().BoolVarP(
		&jsonFormat,
		"json",
		"j",
		false,
		"Format output in JSON",
	)

	ListCmd.Flags().BoolVarP(
		&jsonFormated,
		"json-formatted",
		"f",
		false,
		"Format output in indented JSON",
	)
}
