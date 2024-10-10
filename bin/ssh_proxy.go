package bin

import (
	// "context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/d2jvkpn/socks5-proxy/pkg/proxy"

	"github.com/armon/go-socks5"
	"github.com/d2jvkpn/gotk"
)

func RunSSHProxy(args []string) {
	var (
		fSet    *flag.FlagSet
		config  string
		subkey  string
		network string
		addr    string
		err     error
		logger  *slog.Logger

		client       *proxy.Proxy
		socks5Config *socks5.Config
		listener     net.Listener
		socks5Server *socks5.Server

		errCh    chan error
		shutdown func() error
	)

	// 1.
	shutdown = func() error { return nil }

	fSet = flag.NewFlagSet("proxy", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&config, "config", "configs/local.yaml", "configuration file(yaml)")
	fSet.StringVar(&subkey, "subkey", "ssh", "use subkey of config(yaml)")
	fSet.StringVar(&addr, "addr", ":1081", "socks5 listening address")
	fSet.StringVar(&network, "network", "tcp", "network")

	fSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of proxy:\n")
		fSet.PrintDefaults()
	}

	// fmt.Println("~~~ args:", args)
	if err = fSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ssh exit: %v\n", err)
		os.Exit(1)
		return
	}

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	defer func() {
		if err != nil {
			logger.Error("Exit", "error", err)
			os.Exit(1)
		} else {
			logger.Info("Exit")
		}
	}()

	// 2.
	if client, err = proxy.LoadProxy(config, subkey, nil); err != nil {
		err = fmt.Errorf("Failed to load ssh config: %w", err)
		return
	}

	socks5Config = client.Socks5Config()
	if socks5Server, err = socks5.New(socks5Config); err != nil {
		err = fmt.Errorf("Failed to create SOCKS5 server: %w", err)
		return
	}

	// 3.
	if listener, err = net.Listen(network, addr); err != nil {
		err = fmt.Errorf("Failed to dial %s: %w", addr, err)
		return
	}

	shutdown = func() (err error) {
		err = errors.Join(err, listener.Close())
		err = errors.Join(err, client.Close())
		return err
	}

	// 4.
	errCh = make(chan error, 1)

	go func() {
		var err error

		logger.Info(
			"Starting SOCKS5 proxying",
			"command", "ssh",
			"config", config,
			"address", addr,
			"network", network,
			"authMethods", client.AuthMethods(),
			"socks5User", client.Socks5User,
		)

		err = socks5Server.Serve(listener)
		// accept tcp [::]:1080: use of closed network connection
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			// fmt.Println("####")
			err = nil
		}
		errCh <- err
	}()

	// 5.
	err = gotk.ExitChan(errCh, shutdown)
}
