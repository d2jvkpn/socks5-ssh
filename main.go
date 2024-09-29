package main

import (
	"os"

	"github.com/d2jvkpn/socks5-ssh/bin"
)

func main() {
	bin.RunProxy(os.Args[1:])
}
