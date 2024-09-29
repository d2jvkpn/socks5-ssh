package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/d2jvkpn/socks5-proxy/bin"

	"github.com/d2jvkpn/gotk"
	"github.com/spf13/viper"
)

var (
	//go:embed project.yaml
	_Project []byte
)

func main() {
	var (
		err     error
		project *viper.Viper
		command *gotk.Command
	)

	if project, err = gotk.ProjectFromBytes(_Project); err != nil {
		fmt.Fprintf(os.Stderr, "load project: %s\n", err)
		os.Exit(1)
	}

	command = gotk.NewCommand(project.GetString("app_name"))
	command.Project = project

	command.AddCmd(
		"ssh",
		"socks5 proxy through ssh",
		bin.RunSSHProxy,
	)

	command.AddCmd(
		"show",
		"show config(ssh)",
		func(args []string) {
			if len(args) == 0 {
				return
			}

			switch args[0] {
			case "ssh":
				fmt.Printf("%s\n", command.Project.GetString("ssh_config"))
			default:
			}
		},
	)

	command.Execute(os.Args[1:])
}
