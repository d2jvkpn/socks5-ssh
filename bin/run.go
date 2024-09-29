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

	"github.com/d2jvkpn/socks5-ssh/pkg/proxy"

	"github.com/armon/go-socks5"
	"github.com/d2jvkpn/gotk"
)

func RunProxy(args []string) {
	var (
		flagSet      *flag.FlagSet
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

		errCh    chan error
		shutdown func() error
	)

	// 1.
	shutdown = func() error { return nil }

	flagSet = flag.NewFlagSet("proxy", flag.ContinueOnError) // flag.ExitOnError

	flagSet.StringVar(&config, "config", "configs/local.yaml", "configuration file(yaml)")
	flagSet.StringVar(&subkey, "subkey", "proxy", "use subkey of config(yaml)")
	flagSet.StringVar(&addr, "addr", ":1080", "socks5 listening address")
	flagSet.StringVar(&network, "network", "tcp", "network")

	flagSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of proxy:\n")
		flagSet.PrintDefaults()
	}

	// fmt.Println("~~~ args:", args)
	if err = flagSet.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "proxy exit: %v\n", err)
		os.Exit(1)
		return
	}

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	defer func() {
		if err != nil {
			// fmt.Fprintf(os.Stderr, "\nexit: %s\n", err)
			logger.Error("exit", "error", err)
			os.Exit(1)
		} else {
			logger.Info("exit")
		}
	}()

	// 2.
	if proxyConfig, err = proxy.LoadProxy(config, subkey); err != nil {
		err = fmt.Errorf("Failed to load ssh config: %w", err)
		return
	}

	socks5Config = proxyConfig.Socks5Config()
	if socks5Server, err = socks5.New(socks5Config); err != nil {
		err = fmt.Errorf("Failed to create SOCKS5 server: %w", err)
		return
	}

	// 3.
	if listener, err = net.Listen(network, addr); err != nil {
		err = fmt.Errorf("Failed to dail %s: %w", addr, err)
		return
	}

	shutdown = func() (err error) {
		err = errors.Join(err, listener.Close())
		err = errors.Join(err, proxyConfig.Close())
		return err
	}

	// 4.
	errCh = make(chan error, 1)

	go func() {
		var err error

		logger.Info(
			"Starting SOCKS5 proxyConfig",
			"config", config,
			"address", addr,
			"network", network,
			"authMethods", proxyConfig.AuthMethods(),
			"socks5User", proxyConfig.Socks5User,
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
