package consul

import (
	"embed"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli/v2"
)

//go:embed templates
var ConsulTemplates embed.FS

var Cmd = &cli.Command{
	Name:  "consul",
	Usage: "consul deployment tasks",
	Action: func(c *cli.Context) error {
		f, err := ConsulTemplates.Open("templates/consul.hcl")
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}
