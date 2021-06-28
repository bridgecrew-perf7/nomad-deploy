package consul

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/consul/config"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/consul/deploy"
)

func Up(c *cli.Context) error {
	log.Println("Reading config consul.yaml")
	config, err := config.Load()
	if err != nil {
		return err
	}

	log.Printf("Downloading consul v%s from releases.hashicorp.com\n", config.ConsulVersion)
	deployer, err := deploy.NewDeployer(config)
	if err != nil {
		return err
	}
	defer os.Remove(deployer.ConsulBinPath)

	log.Println("Deploying consul binary to all agents")
	if err := deployer.DeployBinary(); err != nil {
		return err
	}

	log.Println("Deploying service files to all agents")
	if err := deployer.DeployServices(); err != nil {
		return err
	}

	log.Println("Create config directory on all agents")
	if err := deployer.CreateDir("/etc/consul.d/"); err != nil {
		return err
	}

	log.Println("Deploying consul configs to all agents")
	if err := deployer.DeployConsulConfigs(); err != nil {
		return err
	}

	log.Println("Generating TLS certificates")
	certDir, err := deployer.GenerateCertificates()
	if err != nil {
		return err
	}

	log.Println("Deploying TLS certificates")
	if err = deployer.DeployCertificates(certDir); err != nil {
		return err
	}
	if err = os.RemoveAll(certDir); err != nil {
		return err
	}

	log.Println("Creating data directories on all agents")
	if err = deployer.CreateDir("/opt/consul/"); err != nil {
		return err
	}

	log.Println("Enabling and starting consul services on all agents")
	if err = deployer.StartServices(); err != nil {
		return err
	}

	log.Println("Done!")

	if err = deployer.PrintBootstrapTokenInfo(); err != nil {
		return err
	}
	return nil
}
