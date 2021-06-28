package consul

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

func init() {
	Cmd.Subcommands = append(Cmd.Subcommands,
		&cli.Command{
			Name:        "up",
			Description: "Deploy consul cluster",
			Action:      Up,
		},
	)
}

func ReadConfig() (*Config, error) {
	var config Config

	file, err := ioutil.ReadFile("consul.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func DownloadConsulZip(c *Config) (*os.File, error) {
	resp, err := http.Get(fmt.Sprintf("https://releases.hashicorp.com/consul/%s/consul_%s_linux_amd64.zip",
		c.ConsulVersion, c.ConsulVersion))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", fmt.Sprintf("consul%s", c.ConsulVersion))
	if err != nil {
		return nil, err
	}

	io.Copy(file, resp.Body)

	return file, nil
}

func UnzipConsul(zipFile *os.File) (*os.File, error) {
	r, err := zip.OpenReader(zipFile.Name())
	if err != nil {
		return nil, err
	}
	defer r.Close()

	result, err := ioutil.TempFile("", "consul-bin")
	if err != nil {
		return nil, err
	}
	for _, f := range r.File {
		if f.Name == "consul" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			_, err = io.Copy(result, rc)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func CopyConsulToHosts(c *Config, file *os.File) error {
	allHosts := append(c.Clients, c.Servers...)
	for _, host := range allHosts {
		clientConfig, err := auth.PrivateKey(host.User, c.SSHKey, ssh.InsecureIgnoreHostKey())
		if err != nil {
			return err
		}
		client := scp.NewClient(fmt.Sprintf("%s:%d", host.Address, host.SshPort), &clientConfig)
		err = client.Connect()
		if err != nil {
			return err
		}
		defer client.Close()

		err = client.CopyFile(file, "/usr/local/bin/", "0744")
		if err != nil {
			return err
		}
	}
	return nil
}

func Up(c *cli.Context) error {
	log.Println("Reading config consul.yaml")
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	log.Printf("Downloading consul v%s from releases.hashicorp.com\n", config.ConsulVersion)
	consulZip, err := DownloadConsulZip(config)
	defer os.Remove(consulZip.Name())
	if err != nil {
		return err
	}

	log.Println("Unzipping consul archive")
	consulBin, err := UnzipConsul(consulZip)
	if err != nil {
		return err
	}
	defer os.Remove(consulBin.Name())

	log.Println("Copying binary to all hosts")
	if err := CopyConsulToHosts(config, consulBin); err != nil {
		return err
	}

	log.Println("Done!")

	return nil
}
