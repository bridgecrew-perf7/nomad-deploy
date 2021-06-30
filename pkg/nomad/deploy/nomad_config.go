package deploy

import (
	"io/ioutil"
	"os"
	"text/template"

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/ssh"
)

// DeployBaseConfig deploys common between client and server agents
// configuration on hosts
func (c *Nomad) DeployBaseConfig() error {
	tpl, err := template.New("nomad.hcl").ParseFS(c.Templates, "templates/nomad.hcl")
	if err != nil {
		return err
	}
	for _, host := range c.Cfg.AllHosts() {
		tmp, err := ioutil.TempFile("", "nomad.hcl")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())
		tpl.Execute(tmp, map[string]string{
			"DCName":  c.Cfg.DCName,
			"Address": host.Address,
		})
		if err := ssh.Scp(host, c.Cfg, tmp.Name(), "/etc/nomad.d/nomad.hcl"); err != nil {
			return err
		}
	}
	return nil
}

// DeployServerConfig deploys server-only part of configuration
// on all server hosts
func (c *Nomad) DeployServerConfig() error {
	tpl, err := template.New("nomad-server.hcl").ParseFS(templates, "templates/nomad-server.hcl")
	if err != nil {
		return err
	}
	tmp, err := ioutil.TempFile("", "nomad-server.hcl")
	if err != nil {
		return err
	}
	gossipKey, err := c.GenerateGossipKey()
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	tpl.Execute(tmp, map[string]string{
		"GossipKey": gossipKey,
	})
	for _, host := range c.Cfg.Servers {
		if err := ssh.Scp(host, c.Cfg, tmp.Name(), "/etc/nomad.d/nomad-server.hcl"); err != nil {
			return err
		}
	}
	return nil
}

// DeployClientConfig deploys client-only part of configuration
// on all client hosts
func (c *Nomad) DeployClientConfig() error {
	tmp, err := ioutil.TempFile("", "nomad-client.hcl")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	rawConfig, err := templates.ReadFile("templates/nomad-client.hcl")
	if err != nil {
		return err
	}
	_, err = tmp.Write(rawConfig)
	if err != nil {
		return err
	}
	for _, host := range c.Cfg.Clients {
		if err := ssh.Scp(host, c.Cfg, tmp.Name(), "/etc/nomad.d/nomad-client.hcl"); err != nil {
			return err
		}
	}
	return nil
}
