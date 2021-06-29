package deploy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"text/template"
)

func (c *Nomad) DeployNomadConfigs() error {
	commonTpl, err := template.New("nomad.hcl").ParseFS(c.Templates, "templates/nomad.hcl")
	if err != nil {
		return err
	}
	serverFile, err := ioutil.TempFile("", "nomad-server.hcl")
	if err != nil {
		return err
	}
	clientFile, err := ioutil.TempFile("", "nomad-client.hcl")
	if err != nil {
		return err
	}

	parameters := make(map[string]string)
	parameters["DCName"] = c.Cfg.DCName
	if c.Cfg.GossipEnabled {
		log.Println("Generating gossip key")
		parameters["GossipKey"], err = c.GenerateGossipKey()
		if err != nil {
			return err
		}
		log.Println("Generated gossip key: ", parameters["GossipKey"])
	}

	for _, host := range c.Cfg.Clients {
		commonConfig := bytes.Buffer{}
		clientConfig := bytes.Buffer{}
		if err := commonTpl.Execute(&commonConfig, parameters); err != nil {
			return err
		}
		commonFile, err := ioutil.TempFile("", fmt.Sprintf("nomad.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer commonFile.Close()
		serverFile, err := ioutil.TempFile("", fmt.Sprintf("nomad-client.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer serverFile.Close()
		if _, err := io.Copy(commonFile, &commonConfig); err != nil {
			return err
		}
		if _, err := io.Copy(serverFile, &clientConfig); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, commonFile.Name(), "/etc/nomad.d/nomad.hcl"); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, serverFile.Name(), "/etc/nomad.d/nomad-client.hcl"); err != nil {
			return err
		}
	}

	for _, host := range c.Cfg.Servers {
		parameters["Address"] = host.Address
		if c.Cfg.TLSEnabled {
			parameters["CertFile"] = fmt.Sprintf("%s-server-nomad-%d.pem", c.Cfg.DCName, host.Number)
			parameters["KeyFile"] = fmt.Sprintf("%s-server-nomad-%d-key.pem", c.Cfg.DCName, host.Number)
		}

		commonConfig := bytes.Buffer{}
		serverConfig := bytes.Buffer{}
		if err := commonTpl.Execute(&commonConfig, parameters); err != nil {
			return err
		}
		if err := serverTpl.Execute(&serverConfig, parameters); err != nil {
			return err
		}
		commonFile, err := ioutil.TempFile("", fmt.Sprintf("nomad.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer commonFile.Close()
		serverFile, err := ioutil.TempFile("", fmt.Sprintf("nomad-server.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer serverFile.Close()
		if _, err := io.Copy(commonFile, &commonConfig); err != nil {
			return err
		}
		if _, err := io.Copy(serverFile, &serverConfig); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, commonFile.Name(), "/etc/nomad.d/nomad.hcl"); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, serverFile.Name(), "/etc/nomad.d/nomad-server.hcl"); err != nil {
			return err
		}
	}

	return nil
}
