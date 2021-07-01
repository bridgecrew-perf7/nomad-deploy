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

	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/config"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/pkg/ssh"
)

//go:embed templates
var templates embed.FS

type Nomad struct {
	NomadBinPath string
	Cfg          *config.Config
	Templates    fs.FS
}

func NewDeployer(Cfg *config.Config) (*Nomad, error) {
	c := new(Nomad)
	zipFile, err := downloadZip(Cfg)
	if err != nil {
		return nil, err
	}

	c.NomadBinPath, err = unzip(zipFile)
	if err != nil {
		return nil, err
	}
	c.Cfg = Cfg

	c.Templates = templates
	return c, nil
}

func downloadZip(Cfg *config.Config) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://releases.hashicorp.com/nomad/%s/nomad_%s_linux_amd64.zip",
		Cfg.BinaryVersion, Cfg.BinaryVersion))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", fmt.Sprintf("nomad%s", Cfg.BinaryVersion))
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

	result, err := ioutil.TempFile("", "nomad-bin")
	if err != nil {
		return "", err
	}

	defer result.Close()
	for _, f := range r.File {
		if f.Name == "nomad" {
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

func (c *Nomad) DeployBinary() error {
	allHosts := append(c.Cfg.Clients, c.Cfg.Servers...)
	for _, host := range allHosts {
		// check if binary already exists
		bins, err := ssh.Ssh(host, c.Cfg, "ls /usr/local/bin/")
		if err != nil {
			return err
		}
		if strings.Contains(bins, "nomad") {
			continue
		}

		if err := ssh.Scp(host, c.Cfg, c.NomadBinPath, "/usr/local/bin/nomad"); err != nil {
			return err
		}
	}
	return nil
}

// CreateDir creates remote directory with specified path
func (c *Nomad) CreateDir(dirpath string) error {
	for _, host := range append(c.Cfg.Clients, c.Cfg.Servers...) {
		if _, err := ssh.Ssh(host, c.Cfg, fmt.Sprintf("mkdir -p %s", dirpath)); err != nil {
			return err
		}
	}
	return nil
}
