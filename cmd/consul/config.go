package consul

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func init() {
	Cmd.Subcommands = append(Cmd.Subcommands,
		&cli.Command{
			Name:        "config",
			Description: "Generate consul configuration file",
			Action:      generateConfig,
		},
	)
}

const (
	intInput     = iota
	stringInput  = iota
	booleanInput = iota
	hostInput    = iota
)

type Host struct {
	Address   string `yaml:"address"`
	SshPort   int64  `yaml:"sshPort"`
	User      string `yaml:"user"`
	AgentName string `yaml:"agentName"`
}

type Config struct {
	ConsulVersion string `yaml:"version"`
	GossipEnabled bool   `yaml:"gossipEnabled"`
	TLSEnabled    bool   `yaml:"tlsEnabled"`
	Servers       []Host `yaml:"servers"`
	Clients       []Host `yaml:"clients"`
	SSHKey        string `yaml:"sshKey"`
}

func generateConfig(c *cli.Context) error {
	config := survey()
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	os.WriteFile("consul.yaml", configBytes, fs.FileMode(int(0664)))
	log.Println("Config saved in consul.yaml!")
	return nil
}

func question(q string, defaultValue string, t int) interface{} {
	printQuestion := func() {
		switch t {
		case intInput:
			fmt.Printf("[+] %s [%s]: ", q, defaultValue)
		case stringInput:
			fmt.Printf("[+] %s [%s]: ", q, defaultValue)
		case hostInput:
			fmt.Printf("[+] %s [%s]: ", q, defaultValue)
		case booleanInput:
			fmt.Printf("[+] %s (yes/no) [%s]: ", q, defaultValue)
		}
	}

	printQuestion()
	switch t {
	case intInput:
		for {
			var s string
			fmt.Scanln(&s)

			if s == "" {
				result, err := strconv.ParseInt(defaultValue, 0, 64)
				if err != nil {
					log.Fatal(err)
				}
				return result
			}

			result, err := strconv.ParseInt(s, 0, 64)
			if err != nil {
				printQuestion()
			} else {
				return result
			}
		}
	case stringInput:
		var s string
		fmt.Scanln(&s)

		if s == "" {
			return defaultValue
		}
		return s
	case booleanInput:
		parseBool := func(s string) (bool, error) {
			if s == "yes" {
				return true, nil
			} else if s == "no" {
				return false, nil
			} else {
				return false, errors.New("bad boolean parse")
			}
		}
		for {
			var s string
			fmt.Scanln(&s)

			if s == "" {
				result, err := parseBool(defaultValue)
				if err != nil {
					log.Fatal(err)
				}

				return result
			}

			result, err := parseBool(s)
			if err == nil {
				return result
			}
			printQuestion()
		}
	case hostInput:
	QustionLoop:
		for {
			var s string
			fmt.Scanln(&s)

			if s == "" {
				return defaultValue
			}

			octetStrings := strings.Split(s, ".")
			if len(octetStrings) != 4 {
				printQuestion()
				continue
			}

			for _, octet := range octetStrings {
				octetNumber, err := strconv.ParseInt(octet, 0, 64)

				if err != nil {
					printQuestion()
					continue QustionLoop
				}

				if octetNumber < 0 || octetNumber > 255 {
					printQuestion()
					continue QustionLoop
				}
			}
			return s
		}
	}

	return nil
}

func survey() *Config {
	c := new(Config)
	hostsNumber := question("Number of hosts", "1", intInput).(int64)
	for hostNumber := int64(1); hostNumber <= hostsNumber; hostNumber++ {
		address := question(fmt.Sprintf("IP of %d host", hostNumber), "127.0.0.1", hostInput).(string)
		sshPort := question(fmt.Sprintf("SSH port of %d host", hostNumber), "22", intInput).(int64)
		user := question(fmt.Sprintf("Remote user for %d host", hostNumber), "root", stringInput).(string)
		isServer := question(fmt.Sprintf("Is %d host server?", hostNumber), "yes", booleanInput).(bool)

		if isServer {
			c.Servers = append(c.Servers, Host{address, sshPort, user, fmt.Sprintf("server%d", len(c.Servers)+1)})
		} else {
			c.Clients = append(c.Clients, Host{address, sshPort, user, fmt.Sprintf("client%d", len(c.Clients)+1)})
		}
	}
	c.ConsulVersion = question("Consul version", "1.10.0", stringInput).(string)
	c.GossipEnabled = question("Enable gossip encryption?", "yes", booleanInput).(bool)
	c.TLSEnabled = question("Enable tls encryption?", "yes", booleanInput).(bool)
	c.SSHKey = question("Your private SSH key", "~/.ssh/id_rsa", stringInput).(string)

	return c
}
