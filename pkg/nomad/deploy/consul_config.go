package deploy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

func (c *Nomad) DeployConsulConfigs() error {
	commonTpl, err := template.New("consul.hcl").ParseFS(c.Templates, "templates/consul.hcl")
	if err != nil {
		return err
	}
	serverTpl, err := template.New("consul-server.hcl").ParseFS(c.Templates, "templates/consul-server.hcl")
	if err != nil {
		return err
	}
	clientTpl, err := template.New("consul-client.hcl").ParseFS(c.Templates, "templates/consul-client.hcl")
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
	if c.Cfg.ACLEnabled {
		log.Println("Enabling ACL")
		parameters["ACLEnabled"] = "true"
	}
	if c.Cfg.TLSEnabled {
		log.Println("Enabling TLS")
		parameters["CACertFile"] = "consul-agent-ca.pem"
	}

	servers := []string{}
	for _, server := range c.Cfg.Servers {
		servers = append(servers, fmt.Sprintf("\"%s\"", server.Address))
	}

	for _, host := range c.Cfg.Clients {
		parameters["Address"] = host.Address
		if c.Cfg.TLSEnabled {
			parameters["CertFile"] = fmt.Sprintf("%s-client-consul-%d.pem", c.Cfg.DCName, host.Number)
			parameters["KeyFile"] = fmt.Sprintf("%s-client-consul-%d-key.pem", c.Cfg.DCName, host.Number)
		}

		commonConfig := bytes.Buffer{}
		clientConfig := bytes.Buffer{}
		if err := commonTpl.Execute(&commonConfig, parameters); err != nil {
			return err
		}
		err := clientTpl.Execute(&clientConfig, map[string]string{"Servers": "[" + strings.Join(servers, ",") + "]"})
		if err != nil {
			return err
		}
		commonFile, err := ioutil.TempFile("", fmt.Sprintf("consul.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer commonFile.Close()
		serverFile, err := ioutil.TempFile("", fmt.Sprintf("consul-client.hcl%d", host.Number))
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
		if err := Scp(host, c.Cfg, commonFile.Name(), "/etc/consul.d/consul.hcl"); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, serverFile.Name(), "/etc/consul.d/consul-client.hcl"); err != nil {
			return err
		}
	}

	for _, host := range c.Cfg.Servers {
		parameters["Address"] = host.Address
		if c.Cfg.TLSEnabled {
			parameters["CertFile"] = fmt.Sprintf("%s-server-consul-%d.pem", c.Cfg.DCName, host.Number)
			parameters["KeyFile"] = fmt.Sprintf("%s-server-consul-%d-key.pem", c.Cfg.DCName, host.Number)
		}

		commonConfig := bytes.Buffer{}
		serverConfig := bytes.Buffer{}
		if err := commonTpl.Execute(&commonConfig, parameters); err != nil {
			return err
		}
		if err := serverTpl.Execute(&serverConfig, parameters); err != nil {
			return err
		}
		commonFile, err := ioutil.TempFile("", fmt.Sprintf("consul.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer commonFile.Close()
		serverFile, err := ioutil.TempFile("", fmt.Sprintf("consul-server.hcl%d", host.Number))
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
		if err := Scp(host, c.Cfg, commonFile.Name(), "/etc/consul.d/consul.hcl"); err != nil {
			return err
		}
		if err := Scp(host, c.Cfg, serverFile.Name(), "/etc/consul.d/consul-server.hcl"); err != nil {
			return err
		}
	}

	return nil
}
