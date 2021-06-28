package deploy

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

func (c *Consul) DeployServices() error {
	tpl, err := template.New("consul.service").ParseFS(c.Templates, "templates/consul.service")
	if err != nil {
		return err
	}
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		renderedService := bytes.Buffer{}
		err = tpl.Execute(&renderedService, map[string]string{
			"AgentName": host.AgentName,
		})
		if err != nil {
			return err
		}

		tmpFile, err := ioutil.TempFile("", fmt.Sprintf("consul.service%s", host.Address))
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		_, err = tmpFile.Write(renderedService.Bytes())
		if err != nil {
			return err
		}

		if err = Scp(host, c.Cfg, tmpFile.Name(), "/etc/systemd/system/consul.service"); err != nil {
			return err
		}
	}
	return nil
}

func (c *Consul) StartServices() error {
	for _, host := range append(c.Cfg.Servers, c.Cfg.Clients...) {
		if _, err := Ssh(host, c.Cfg, "systemctl enable consul.service"); err != nil {
			return err
		}
		if _, err := Ssh(host, c.Cfg, "systemctl start consul.service"); err != nil {
			return err
		}
	}
	return nil
}
