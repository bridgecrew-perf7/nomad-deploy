package consul

import (
	"log"

	"github.com/urfave/cli/v2"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/consul/config"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/consul/deploy"
)

func Remove(c *cli.Context) error {
	log.Println("Reading config consul.yaml")
	config, err := config.Load()
	if err != nil {
		return err
	}

	deployer := deploy.Consul{Cfg: config}

	log.Println("Stopping and deleting services")
	if err := deployer.DeleteServices(); err != nil {
		return err
	}

	log.Println("Deleting config directory")
	if err := deployer.DeleteConfigs(); err != nil {
		return err
	}

	log.Println("Deleting data directory")
	if err := deployer.DeleteData(); err != nil {
		return err
	}

	log.Println("Done!")
	return nil
}
