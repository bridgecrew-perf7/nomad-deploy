package cmd

import (
	"github.com/urfave/cli/v2"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/cmd/consul"
	"gitlab.gs-labs.tv/casdevops/nomad-deploy/cmd/nomad"
)

var App = &cli.App{
	Name:  "nomad-deploy",
	Usage: "Deploy consul and nomad with ease",
	Commands: []*cli.Command{
		consul.Cmd,
		nomad.Cmd,
	},
}
