package main

import (
	// "fmt"
	"os"

	"github.com/d2jvkpn/gotk"
	"github.com/d2jvkpn/socks5-ssh/bin"
)

func main() {
	var command *gotk.Command

	command = gotk.NewCommand("socks5-ssh")

	command.AddCmd(
		"proxy",
		"socks5 proxy through ssh",
		bin.RunProxy,
	)

	command.Execute(os.Args[1:])
}
