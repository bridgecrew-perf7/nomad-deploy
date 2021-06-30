package deploy

import "gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/ssh"

func (c *Consul) DeleteServices() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(
			host,
			c.Cfg,
			"bash -c \"systemctl stop consul; systemctl disable consul; rm -f /etc/systemd/system/consul.service\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Consul) DeleteConfigs() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(host, c.Cfg, "bash -c \"rm -rf /etc/consul.d\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Consul) DeleteData() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := ssh.Ssh(host, c.Cfg, "bash -c \"rm -rf /opt/consul\"")
		if err != nil {
			return err
		}
	}
	return nil
}
