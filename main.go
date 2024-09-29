package main

import (
	"os"

	"github.com/d2jvkpn/socks5-ssh/bin/proxyCmd"
)

func main() {
	proxyCmd.Run(os.Args[1:])
}
