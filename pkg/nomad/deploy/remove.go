package deploy

func (c *Nomad) DeleteServices() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(
			host,
			c.Cfg,
			"bash -c \"systemctl stop consul; systemctl disable consul; rm -f /etc/systemd/system/consul.service\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Nomad) DeleteConfigs() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(host, c.Cfg, "bash -c \"rm -rf /etc/consul.d\"")
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Nomad) DeleteData() error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		_, err := Ssh(host, c.Cfg, "bash -c \"rm -rf /opt/consul\"")
		if err != nil {
			return err
		}
	}
	return nil
}
