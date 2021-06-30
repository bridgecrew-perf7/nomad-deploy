package nomad

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/config"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/nomad/deploy"
)

func Up(c *cli.Context) error {
	log.Println("Reading config nomad.yaml")
	config, err := config.Load()
	if err != nil {
		return err
	}

	log.Printf("Downloading nomad v%s from releases.hashicorp.com\n", config.BinaryVersion)
	deployer, err := deploy.NewDeployer(config)
	if err != nil {
		return err
	}
	defer os.Remove(deployer.NomadBinPath)

	log.Println("Deploying nomad binary to all agents")
	if err := deployer.DeployBinary(); err != nil {
		return err
	}

	log.Println("Deploying systemd service file to all agents")
	if err := deployer.DeploySystemd(); err != nil {
		return err
	}

	log.Println("Create config directory on all agents")
	if err := deployer.CreateDir("/etc/nomad.d/"); err != nil {
		return err
	}

	log.Println("Deploying common nomad config to all agents")
	if err := deployer.DeployBaseConfig(); err != nil {
		return err
	}

	log.Println("Deploying server config to all agents")
	if err := deployer.DeployServerConfig(); err != nil {
		return err
	}

	log.Println("Deploying client config to all agents")
	if err := deployer.DeployClientConfig(); err != nil {
		return err
	}

	log.Println("Creating data directories on all agents")
	if err = deployer.CreateDir("/opt/nomad/"); err != nil {
		return err
	}

	log.Println("Enabling and starting nomad systemd services on all agents")
	if err = deployer.StartSystemd(); err != nil {
		return err
	}

	log.Println("Done!")

	return nil
}
