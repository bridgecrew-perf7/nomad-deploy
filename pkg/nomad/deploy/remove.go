package deploy

import "gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/ssh"

// DeleteSystemd deletes systemd service file
func (c *Nomad) DeleteSystemd() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(
			host,
			c.Cfg,
			"bash -c \"systemctl stop nomad; systemctl disable nomad; rm -f /etc/systemd/system/nomad.service\"")
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteConfigs deletes configuration directory
func (c *Nomad) DeleteConfigs() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(host, c.Cfg, "bash -c \"rm -rf /etc/nomad.d\"")
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteData deletes data directory
func (c *Nomad) DeleteData() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(host, c.Cfg, "bash -c \"rm -rf /opt/nomad\"")
		if err != nil {
			return err
		}
	}
	return nil
}
