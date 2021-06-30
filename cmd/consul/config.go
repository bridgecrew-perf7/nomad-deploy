package consul

import (
	"github.com/urfave/cli/v2"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/config"
)

func GenerateConfig(c *cli.Context) error {
	config := config.Survey()
	return config.Save()
}
