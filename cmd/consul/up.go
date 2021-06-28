package consul

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"text/template"

	"github.com/urfave/cli/v2"
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
		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			file.Name(),
			fmt.Sprintf("%s@%s:/usr/local/bin", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err := scpCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func DeploySeriveDescription(c *Config) error {
	tpl, err := template.New("consul.service").ParseFS(ConsulTemplates, "templates/consul.service")
	if err != nil {
		return err
	}

	for _, host := range append(c.Clients, c.Servers...) {
		renderedService := bytes.Buffer{}
		err = tpl.Execute(&renderedService, map[string]string{
			"AgentName": host.AgentName,
		})
		if err != nil {
			return err
		}

		tempFile, err := ioutil.TempFile("", fmt.Sprintf("consul.service%s", host.Address))
		if err != nil {
			return err
		}
		defer tempFile.Close()

		_, err = tempFile.Write(renderedService.Bytes())
		if err != nil {
			return err
		}

		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			tempFile.Name(),
			fmt.Sprintf("%s@%s:/etc/systemd/system/consul.service", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err := scpCmd.Run(); err != nil {
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

	log.Println("Deploying consul binary to all agents")
	if err := CopyConsulToHosts(config, consulBin); err != nil {
		return err
	}

	log.Println("Deploying service description to all agents")
	if err := DeploySeriveDescription(config); err != nil {
		return err
	}

	// log.Println("Generating consul client certificates")
	// log.Println("Deploying consul client certificates")
	// log.Println("Generating consul server certificates")
	// log.Println("Deploying consul server certificates")
	// log.Println("Deploying config to client agents")
	// log.Println("Deploying config to server agents")
	// log.Println("Creating data directories on all agents")
	// log.Println("Enabling and starting consul services on all agents")

	log.Println("Done!")
	return nil
}
