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

	//go:embed deployments/compose.socks5_ssh.yaml
	_ComposeSSH []byte

	//go:embed deployments/compose.socks5_vpn.yaml
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
		"config", "show config(ssh_proxy, socks5_proxy)",
		func(args []string) {
			errMsg := "Subcommand is required: ssh_proxy | socks5_proxy\n"

			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
				return
			}

			switch args[0] {
			case "ssh_proxy", "socks5_proxy":
				fmt.Printf("%s\n", command.Project.GetString(args[0]))
			default:
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
			}
		},
	)

	command.AddCmd(
		"script", "show script(proxy_pac, compose_ssh, compose_vpn)",
		func(args []string) {
			errMsg := "Subcommand is required: proxy_pac | compose_ssh | compose_vpn\n"

			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
				return
			}

			switch args[0] {
			case "proxy_pac":
				fmt.Printf("%s\n", _ProxyPac)
			case "compose_ssh":
				fmt.Printf("%s\n", _ComposeSSH)
			case "compose_vpn":
				fmt.Printf("%s\n", _ComposeVPN)
			default:
				fmt.Fprintf(os.Stderr, errMsg)
				os.Exit(1)
			}
		},
	)

	command.AddCmd(
		"ssh", "socks5 proxy through ssh",
		bin.RunProxyXSSH,
	)

	command.AddCmd(
		"server", "socks5 proxy server",
		bin.RunProxyServer,
	)

	command.AddCmd(
		"file_server", "http file server for pac files",
		bin.RunFileServer,
	)

	command.AddCmd(
		"test", "test socks5 proxy",
		bin.TestProxy,
	)

	command.Execute(os.Args[1:])
}
