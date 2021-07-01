package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/ssh"
)

// GenerateGossipKey generates random 20-byte key using
// nomad binary
func (c *Nomad) GenerateGossipKey() (string, error) {
	key := bytes.Buffer{}
	cmd := exec.Command(c.NomadBinPath, "operator", "keygen")
	cmd.Stderr = os.Stderr
	cmd.Stdout = &key
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(key.String()), nil
}

func (c *Nomad) GenerateCertificates() (string, error) {
	tempDir, err := ioutil.TempDir("", "nomad-cert")
	if err != nil {
		return "", err
	}
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if err = os.Chdir(tempDir); err != nil {
		return "", err
	}

	//create CA
	createCaCmd := exec.Command(c.NomadBinPath, "tls", "ca", "create")
	createCaCmd.Stderr = os.Stderr
	if err = createCaCmd.Run(); err != nil {
		return "", err
	}

	//create client certs
	for range c.Cfg.Clients {
		createClientCert := exec.Command(c.NomadBinPath, "tls", "cert", "create", "-client",
			fmt.Sprintf("-dc=%s", c.Cfg.DCName))
		createClientCert.Stderr = os.Stderr
		if err = createClientCert.Run(); err != nil {
			return "", err
		}
	}

	//create server certs
	for range c.Cfg.Servers {
		createServerCert := exec.Command(c.NomadBinPath, "tls", "cert", "create", "-server",
			fmt.Sprintf("-dc=%s", c.Cfg.DCName))
		createServerCert.Stderr = os.Stderr
		if err = createServerCert.Run(); err != nil {
			return "", err
		}
	}

	if err = os.Chdir(currentDir); err != nil {
		return "", err
	}
	return tempDir, nil
}

func (c *Nomad) DeployCertificates(certsDir string) error {
	findFiles := func(dir string, pattern string) ([]string, error) {
		files := []string{}
		err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
			if strings.Contains(f.Name(), pattern) {
				files = append(files, path)
			}
			return nil
		})
		return files, err
	}
	for _, host := range c.Cfg.Clients {
		certs, err := findFiles(certsDir, "client")
		if err != nil {
			return err
		}
		for _, cert := range append(certs, filepath.Join(certsDir, "nomad-agent-ca.pem")) {
			err = ssh.Scp(host, c.Cfg, cert, "/etc/nomad.d/")
			if err != nil {
				return err
			}
		}
	}
	for _, host := range c.Cfg.Servers {
		certs, err := findFiles(certsDir, "server")
		if err != nil {
			return err
		}
		for _, cert := range append(certs, filepath.Join(certsDir, "nomad-agent-ca.pem")) {
			err = ssh.Scp(host, c.Cfg, cert, "/etc/nomad.d/")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Nomad) PrintBootstrapTokenInfo() error {
	output, err := ssh.Ssh(c.Cfg.Servers[0], c.Cfg, "nomad acl bootstrap")
	if err != nil {
		return err
	}
	fmt.Println("Your bootstrapped ACL token:")
	fmt.Println(output)
	return nil
}
