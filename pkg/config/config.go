package config

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Host struct {
	Address   string `yaml:"address"`
	SshPort   int64  `yaml:"sshPort"`
	User      string `yaml:"user"`
	AgentName string `yaml:"agentName"`
	Number    int    `yaml:"number"`
}

type Config struct {
	BinaryVersion string `yaml:"version"`
	GossipEnabled bool   `yaml:"gossipEnabled"`
	ACLEnabled    bool   `yaml:"aclEnabled"`
	TLSEnabled    bool   `yaml:"tlsEnabled"`
	Servers       []Host `yaml:"servers"`
	Clients       []Host `yaml:"clients"`
	SSHKey        string `yaml:"sshKey"`
	DCName        string `yaml:"dcName"`
}

func (c *Config) Save() error {
	configBytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err = os.WriteFile("consul.yaml", configBytes, fs.FileMode(int(0664))); err != nil {
		return err
	}
	log.Println("Config saved in consul.yaml!")
	return nil
}

func Load() (*Config, error) {
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

func (c *Config) AllHosts() []Host {
	return append(c.Servers, c.Clients...)
}
