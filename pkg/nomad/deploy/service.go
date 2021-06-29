package deploy

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

func (c *Nomad) DeployServices() error {
	tpl, err := template.New("nomad.service").ParseFS(c.Templates, "templates/nomad.service")
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

		tmpFile, err := ioutil.TempFile("", fmt.Sprintf("nomad.service%s", host.Address))
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		_, err = tmpFile.Write(renderedService.Bytes())
		if err != nil {
			return err
		}

		if err = Scp(host, c.Cfg, tmpFile.Name(), "/etc/systemd/system/nomad.service"); err != nil {
			return err
		}
	}
	return nil
}

func (c *Nomad) StartServices() error {
	for _, host := range append(c.Cfg.Servers, c.Cfg.Clients...) {
		_, err := Ssh(host, c.Cfg,
			"systemctl enable nomad.service; systemctl start nomad.service")
		if err != nil {
			return err
		}
	}
	return nil
}
