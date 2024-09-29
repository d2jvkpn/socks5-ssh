package main

import (
	// "fmt"
	"os"

	"github.com/d2jvkpn/gotk"
	"github.com/d2jvkpn/socks5-proxy/bin"
)

func main() {
	var command *gotk.Command

	command = gotk.NewCommand("socks5")

	command.AddCmd(
		"ssh",
		"socks5 proxy through ssh",
		bin.RunSSHProxy,
	)

	command.Execute(os.Args[1:])
}
