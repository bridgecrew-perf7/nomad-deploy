package deploy

func (c *Nomad) DeleteServices() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(
			host,
			c.Cfg,
			"bash -c \"systemctl stop nomad; systemctl disable nomad; rm -f /etc/systemd/system/nomad.service\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Nomad) DeleteConfigs() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(host, c.Cfg, "bash -c \"rm -rf /etc/nomad.d\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Nomad) DeleteData() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(host, c.Cfg, "bash -c \"rm -rf /opt/nomad\"")
		if err != nil {
			return err
		}
	}
	return nil
}
