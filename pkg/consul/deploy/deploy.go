package deploy

import (
	"archive/zip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/consul/config"
)

//go:embed templates
var templates embed.FS

type Consul struct {
	ConsulBinPath string
	Cfg           *config.Config
	Templates     fs.FS
}

func NewDeployer(Cfg *config.Config) (*Consul, error) {
	c := new(Consul)
	zipFile, err := downloadZip(Cfg)
	if err != nil {
		return nil, err
	}

	c.ConsulBinPath, err = unzip(zipFile)
	if err != nil {
		return nil, err
	}
	c.Cfg = Cfg

	c.Templates = templates
	return c, nil
}

func downloadZip(Cfg *config.Config) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://releases.hashicorp.com/consul/%s/consul_%s_linux_amd64.zip",
		Cfg.ConsulVersion, Cfg.ConsulVersion))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", fmt.Sprintf("consul%s", Cfg.ConsulVersion))
	if err != nil {
		return "", err
	}
	defer file.Close()

	io.Copy(file, resp.Body)
	return file.Name(), nil
}

func unzip(zipFile string) (string, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", err
	}
	defer r.Close()

	result, err := ioutil.TempFile("", "consul-bin")
	if err != nil {
		return "", err
	}

	defer result.Close()
	for _, f := range r.File {
		if f.Name == "consul" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()
			_, err = io.Copy(result, rc)
			if err != nil {
				return "", err
			}
		}
	}
	err = result.Chmod(0755)
	if err != nil {
		return "", err
	}
	return result.Name(), nil
}

func (c *Consul) DeployBinary() error {
	allHosts := append(c.Cfg.Clients, c.Cfg.Servers...)
	for _, host := range allHosts {
		// check if binary already exists
		bins, err := Ssh(host, c.Cfg, "ls /usr/local/bin/")
		if err != nil {
			return err
		}
		if strings.Contains(bins, "consul") {
			continue
		}

		if err := Scp(host, c.Cfg, c.ConsulBinPath, "/usr/local/bin/consul"); err != nil {
			return err
		}
	}
	return nil
}
