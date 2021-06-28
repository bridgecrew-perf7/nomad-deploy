package consul

import (
	"log"

	"github.com/urfave/cli/v2"
)

func init() {
	Cmd.Subcommands = append(Cmd.Subcommands,
		&cli.Command{
			Name:        "remove",
			Description: "Destroy consul cluster",
			Action:      Remove,
		},
	)
}

func Remove(c *cli.Context) error {
	log.Println("Reading config config.yaml")

	return nil
}
