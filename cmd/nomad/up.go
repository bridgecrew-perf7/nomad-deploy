package nomad

import "github.com/urfave/cli/v2"

func init() {
	Cmd.Subcommands = append(Cmd.Subcommands,
		&cli.Command{
			Name:        "up",
			Description: "Deploy nomad cluster",
		},
	)
}
