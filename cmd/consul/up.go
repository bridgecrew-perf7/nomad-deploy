package consul

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	defer result.Close()
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
	err = result.Chmod(0755)
	if err != nil {
		return nil, err
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
			fmt.Sprintf("%s@%s:/usr/local/bin/consul", host.User, host.Address))
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

		tmpFile, err := ioutil.TempFile("", fmt.Sprintf("consul.service%s", host.Address))
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		_, err = tmpFile.Write(renderedService.Bytes())
		if err != nil {
			return err
		}

		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			tmpFile.Name(),
			fmt.Sprintf("%s@%s:/etc/systemd/system/consul.service", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err := scpCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func GenerateGossipKey(consulBin *os.File) (string, error) {
	key := bytes.Buffer{}
	cmd := exec.Command(consulBin.Name(), "keygen")
	cmd.Stderr = os.Stderr
	cmd.Stdout = &key
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(key.String()), nil
}

func CreateConfigDir(c *Config) error {
	for _, host := range append(c.Clients, c.Servers...) {
		mkdirCmd := exec.Command("ssh",
			"-p", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			fmt.Sprintf("%s@%s", host.User, host.Address),
			"mkdir -p /etc/consul.d")
		mkdirCmd.Stderr = os.Stderr
		if err := mkdirCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func CreateDataDir(c *Config) error {
	for _, host := range append(c.Clients, c.Servers...) {
		mkdirCmd := exec.Command("ssh",
			"-p", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			fmt.Sprintf("%s@%s", host.User, host.Address),
			"mkdir -p /opt/consul")
		mkdirCmd.Stderr = os.Stderr
		if err := mkdirCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func DeployConfigs(c *Config, consulBin *os.File) error {
	tpl, err := template.New("consul.hcl").ParseFS(ConsulTemplates, "templates/consul.hcl")
	if err != nil {
		return err
	}

	parameters := make(map[string]string)
	parameters["DCName"] = c.DCName
	if c.GossipEnabled {
		log.Println("Generating gossip key")
		parameters["GossipKey"], err = GenerateGossipKey(consulBin)
		if err != nil {
			return err
		}
		log.Println("Generated gossip key: ", parameters["GossipKey"])
	}
	if c.ACLEnabled {
		log.Println("Enabling ACL")
		parameters["ACLEnabled"] = "true"
	}
	if c.TLSEnabled {
		log.Println("Enabling TLS")
		parameters["CACertFile"] = "consul-agent-ca.pem"
	}

	deployConsulConfig := func(host Host, hostType string) error {
		renderedService := bytes.Buffer{}
		parameters["Address"] = host.Address
		if c.TLSEnabled {
			parameters["CertFile"] = fmt.Sprintf("%s-%s-consul-%d.pem", c.DCName, hostType, host.Number)
			parameters["KeyFile"] = fmt.Sprintf("%s-%s-consul-%d-key.pem", c.DCName, hostType, host.Number)
		}
		err = tpl.Execute(&renderedService, parameters)
		if err != nil {
			return err
		}

		tmpFile, err := ioutil.TempFile("", fmt.Sprintf("consul.hcl%d", host.Number))
		if err != nil {
			return err
		}
		defer tmpFile.Close()

		_, err = tmpFile.Write(renderedService.Bytes())
		if err != nil {
			return err
		}

		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			tmpFile.Name(),
			fmt.Sprintf("%s@%s:/etc/consul.d/consul.hcl", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err := scpCmd.Run(); err != nil {
			return err
		}
		return nil
	}

	for _, host := range c.Clients {
		if err := deployConsulConfig(host, "client"); err != nil {
			return err
		}
	}
	for _, host := range c.Servers {
		if err := deployConsulConfig(host, "server"); err != nil {
			return err
		}
	}

	return nil
}

func GenerateCertificates(c *Config, consulBin *os.File) (string, error) {
	tempDir, err := ioutil.TempDir("", "consul-cert")
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
	createCaCmd := exec.Command(consulBin.Name(), "tls", "ca", "create")
	createCaCmd.Stderr = os.Stderr
	if err = createCaCmd.Run(); err != nil {
		return "", err
	}

	//create client certs
	for range c.Clients {
		createClientCert := exec.Command(consulBin.Name(), "tls", "cert", "create", "-client",
			fmt.Sprintf("-dc=%s", c.DCName))
		createClientCert.Stderr = os.Stderr
		if err = createClientCert.Run(); err != nil {
			return "", err
		}
	}

	//create server certs
	for range c.Servers {
		createServerCert := exec.Command(consulBin.Name(), "tls", "cert", "create", "-server",
			fmt.Sprintf("-dc=%s", c.DCName))
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

func DeployCertificates(c *Config, certDir string) error {
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

	deployCerts := func(host Host, hostType string) error {
		certs, err := findFiles(certDir, hostType)
		if err != nil {
			return err
		}

		params := []string{
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			filepath.Join(certDir, "consul-agent-ca.pem"),
		}
		params = append(params, certs...)
		params = append(params, fmt.Sprintf("%s@%s:/etc/consul.d/", host.User, host.Address))

		scpCertsCmd := exec.Command("scp", params...)
		scpCertsCmd.Stderr = os.Stderr
		if err := scpCertsCmd.Run(); err != nil {
			return err
		}
		return nil
	}
	for _, host := range c.Clients {
		if err := deployCerts(host, "client"); err != nil {
			return err
		}
	}
	for _, host := range c.Servers {
		if err := deployCerts(host, "server"); err != nil {
			return err
		}
	}

	return nil
}

func DeployServerConfigs(c *Config) error {
	file, err := ConsulTemplates.Open("templates/consul-server.hcl")
	if err != nil {
		return err
	}
	defer file.Close()
	tmpFile, err := ioutil.TempFile("", "consul-server.hcl")
	if err != nil {
		return err
	}

	io.Copy(tmpFile, file)
	defer os.Remove(tmpFile.Name())

	for _, host := range c.Servers {
		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			tmpFile.Name(),
			fmt.Sprintf("%s@%s:/etc/consul.d/consul-server.hcl", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err = scpCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func DeployClientConfigs(c *Config) error {
	tpl, err := template.New("consul-client.hcl").ParseFS(ConsulTemplates, "templates/consul-client.hcl")
	if err != nil {
		return err
	}

	servers := []string{}
	for _, server := range c.Servers {
		servers = append(servers, fmt.Sprintf("\"%s\"", server.Address))
	}
	tmpFile, err := ioutil.TempFile("", "consul-client.hcl")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	tpl.Execute(tmpFile, map[string]string{
		"Servers": "[" + strings.Join(servers, ",") + "]",
	})

	for _, host := range c.Clients {
		scpCmd := exec.Command("scp",
			"-P", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			tmpFile.Name(),
			fmt.Sprintf("%s@%s:/etc/consul.d/consul-client.hcl", host.User, host.Address))
		scpCmd.Stderr = os.Stderr
		if err = scpCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func StartSystemd(c *Config) error {
	for _, host := range append(c.Servers, c.Clients...) {
		systemctlEnableCmd := exec.Command("ssh",
			"-p", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			fmt.Sprintf("%s@%s", host.User, host.Address),
			"systemctl enable consul.service")
		systemctlStartCmd := exec.Command("ssh",
			"-p", strconv.Itoa(int(host.SshPort)),
			"-i", c.SSHKey,
			fmt.Sprintf("%s@%s", host.User, host.Address),
			"systemctl start consul.service")
		systemctlEnableCmd.Stderr = os.Stderr
		systemctlStartCmd.Stderr = os.Stderr
		if err := systemctlEnableCmd.Run(); err != nil {
			return err
		}
		if err := systemctlStartCmd.Run(); err != nil {
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

	log.Println("Create config directory on all agents")
	if err := CreateConfigDir(config); err != nil {
		return err
	}

	log.Println("Deploying common configs to all agents")
	if err := DeployConfigs(config, consulBin); err != nil {
		return err
	}

	log.Println("Deploying server-specific config to server agents")
	if err := DeployServerConfigs(config); err != nil {
		return err
	}

	log.Println("Deploying client-specific config to client agents")
	if err := DeployClientConfigs(config); err != nil {
		return err
	}

	log.Println("Generating TLS certificates")
	certDir, err := GenerateCertificates(config, consulBin)
	if err != nil {
		return err
	}

	log.Println("Deploying TLS certificates")
	if err = DeployCertificates(config, certDir); err != nil {
		return err
	}
	if err = os.RemoveAll(certDir); err != nil {
		return err
	}

	log.Println("Creating data directories on all agents")
	if err = CreateDataDir(config); err != nil {
		return err
	}

	log.Println("Enabling and starting consul services on all agents")
	if err = StartSystemd(config); err != nil {
		return err
	}

	// data, err := yaml.Marshal(config)
	// if err != nil {
	// 	return err
	// }
	// os.WriteFile("consul.state.yaml", data, fs.FileMode(int(0664)))

	log.Println("Done!")
	return nil
}
