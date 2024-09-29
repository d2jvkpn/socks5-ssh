package main

import (
	// "context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/d2jvkpn/socks5-ssh/pkg/proxy"

	"github.com/armon/go-socks5"
)

func main() {
	var (
		debug        bool
		config       string
		subkey       string
		network      string
		addr         string
		err          error
		logger       *slog.Logger
		proxyConfig  *proxy.Proxy
		socks5Config *socks5.Config
		listener     net.Listener
		socks5Server *socks5.Server

		count int
		errCh chan error
		sigCh chan os.Signal
	)

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	flag.StringVar(&config, "config", "configs/local.yaml", "configuration file(yaml)")
	flag.StringVar(&subkey, "subkey", "proxy", "use subkey of config(yaml)")
	flag.StringVar(&addr, "addr", ":1080", "socks5 listening address")
	flag.StringVar(&network, "network", "tcp", "network")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.Parse()

	defer func() {
		err = errors.Join(err, proxyConfig.Close())

		if err != nil {
			fmt.Fprintf(os.Stderr, "\nEXIT: %s\n", err)
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "\nEXIT: 0\n")
		}
	}()

	if proxyConfig, err = proxy.LoadProxy(config, subkey); err != nil {
		err = fmt.Errorf("Failed to load ssh config: %w", err)
		return
	}

	socks5Config = proxyConfig.Socks5Config()
	if socks5Server, err = socks5.New(socks5Config); err != nil {
		err = fmt.Errorf("Failed to create SOCKS5 server: %w", err)
		return
	}

	if listener, err = net.Listen(network, addr); err != nil {
		err = fmt.Errorf("Failed to dail %s: %w", addr, err)
		return
	}

	errCh = make(chan error, 1)
	sigCh = make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		var err error

		logger.Info(
			"Starting SOCKS5 proxyConfig",
			"config", config,
			"address", addr,
			"debug", debug,
			"network", network,
			"authMethods", proxyConfig.AuthMethods(),
			"socks5User", proxyConfig.Socks5User,
		)

		err = socks5Server.Serve(listener)
		errCh <- err
	}()

	count = cap(errCh)

	syncErr := func(count int) {
		for i := 0; i < count; i++ {
			err = errors.Join(err, <-errCh)
		}
	}

	select {
	case e := <-errCh:
		err = errors.Join(err, e)
		count -= 1
	case <-sigCh:
		err = errors.Join(err, listener.Close())
	}

	syncErr(count)
}
