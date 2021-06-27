package consul

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

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

func DownloadConsul(c *Config) (*os.File, error) {
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

	return file, nil
}

func CopyConsulToHosts(c *Config, file *os.File) error {
	allHosts := append(c.Clients, c.Servers...)
	unzipCmd := exec.Command("unzip",
		file.Name(),
		"-d", "/tmp")
	if err := unzipCmd.Run(); err != nil {
		return err
	}

	for _, host := range allHosts {
		copyCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"/tmp/consul",
			fmt.Sprintf("%s@%s:/home/%s/", host.User, host.Address, host.User))
		if err := copyCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func Up(c *cli.Context) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}

	consulZip, err := DownloadConsul(config)
	if err != nil {
		return err
	}

	os.Remove(consulZip.Name())
	return nil
}
