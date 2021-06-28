package consul

import (
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:  "consul",
	Usage: "consul deployment tasks",
	Subcommands: []*cli.Command{
		{
			Name:        "up",
			Description: "Deploy consul cluster",
			Action:      Up,
		},
		{
			Name:        "config",
			Description: "Generate config via interactive survey",
			Action:      GenerateConfig,
		},
		{
			Name:        "remove",
			Description: "Clear all consul traces",
			Action:      Remove,
		},
	},
}
