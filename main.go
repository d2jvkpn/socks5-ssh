package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/d2jvkpn/socks5-proxy/bin"

	"github.com/d2jvkpn/gotk"
	// "github.com/spf13/viper"
)

var (
	//go:embed project.yaml
	_Project []byte
)

func main() {
	var (
		err     error
		command *gotk.Command
	)

	command = gotk.NewCommand("socks5")
	if command.Project, err = gotk.ProjectFromBytes(_Project); err != nil {
		fmt.Fprintf(os.Stderr, "load project: %s\n", err)
		os.Exit(1)
	}

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
