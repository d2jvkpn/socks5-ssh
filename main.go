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

	//go:embed deployments/compose.socks5-ssh.yaml
	_ComposeSSH []byte

	//go:embed deployments/compose.socks5-vpn.yaml
	_ComposeVPN []byte

	//go:embed deployments/proxy.pac
	_ProxyPac []byte
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
		"config", "show config(socks5_ssh, ssh_config, supervisor_vpn, supervisor_ssh)",
		func(args []string) {
			errMsg := "Subcommand is required: " +
				"socks5_ssh | ssh_config | supervisor_vpn | supervisor_ssh\n"

			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
				return
			}

			switch args[0] {
			case "socks5_ssh", "ssh_config", "supervisor_vpn", "supervisor_ssh":
				fmt.Printf("%s\n", command.Project.GetString(args[0]))
			default:
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
			}
		},
	)

	command.AddCmd(
		"script", "show script(pac, ssh, vpn)",
		func(args []string) {
			errMsg := "Subcommand is required: pac | ssh | vpn\n"

			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
				return
			}

			switch args[0] {
			case "pac":
				fmt.Printf("%s\n", _ProxyPac)
			case "ssh":
				fmt.Printf("%s\n", _ComposeSSH)
			case "vpn":
				fmt.Printf("%s\n", _ComposeVPN)
			default:
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
			}
		},
	)

	command.AddCmd(
		"ssh", "run a socks5 proxying server through ssh",
		bin.RunSocks5SSH,
	)

	command.AddCmd(
		"server", "run a socks5 proxying server",
		bin.RunSocks5Server,
	)

	command.AddCmd(
		"file_server", "run a http file server(provide pac files for browsers)",
		bin.RunFileServer,
	)

	command.AddCmd(
		"test", "test socks5 proxy",
		bin.TestProxy,
	)

	command.Execute(os.Args[1:])
}
