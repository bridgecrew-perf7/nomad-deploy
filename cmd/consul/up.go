package consul

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

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
	privateKey, err := ioutil.ReadFile(c.SSHKey)
	if err != nil {
		return err
	}
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return err
	}
	for _, host := range append(c.Clients, c.Servers...) {
		sshConfig := &ssh.ClientConfig{
			User: host.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", host.Address, host.SshPort), sshConfig)
		if err != nil {
			return err
		}

		session, err := client.NewSession()
		if err != nil {
			return err
		}
		defer session.Close()

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
