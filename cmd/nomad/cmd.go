package nomad

import (
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:  "nomad",
	Usage: "nomad deployment tasks",
	Subcommands: []*cli.Command{
		{
			Name:        "up",
			Description: "Deploy nomad cluster",
			Action:      Up,
		},
		{
			Name:        "config",
			Description: "Generate config via interactive survey",
			Action:      GenerateConfig,
		},
		{
			Name:        "remove",
			Description: "Clear all nomad traces",
			Action:      Remove,
		},
	},
}
