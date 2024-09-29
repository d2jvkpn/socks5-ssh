package proxyCmd

import (
	// "context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/d2jvkpn/socks5-ssh/pkg/proxy"

	"github.com/armon/go-socks5"
)

func Run(args []string) {
	var (
		flagSet      *flag.FlagSet
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

	// 1.
	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	defer func() {
		err = errors.Join(err, proxyConfig.Close())

		if err != nil {
			// fmt.Fprintf(os.Stderr, "\nexit: %s\n", err)
			logger.Error("exit", "error", err)
			os.Exit(1)
		} else {
			logger.Info("exit")
		}
	}()

	flagSet = flag.NewFlagSet("run_proxy", flag.ExitOnError) // flag.ContinueOnError

	flagSet.StringVar(&config, "config", "configs/local.yaml", "configuration file(yaml)")
	flagSet.StringVar(&subkey, "subkey", "proxy", "use subkey of config(yaml)")
	flagSet.StringVar(&addr, "addr", ":1080", "socks5 listening address")
	flagSet.StringVar(&network, "network", "tcp", "network")
	flagSet.BoolVar(&debug, "debug", false, "enable debug")

	flagSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of %s:\n", "proxy")
		flagSet.PrintDefaults()
	}

	// fmt.Println("~~~ args:", args)
	if err = flagSet.Parse(args); err != nil {
		return
	}

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

	// 4.
	errCh = make(chan error, 1)
	count = cap(errCh)

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
		// accept tcp [::]:1080: use of closed network connection
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			// fmt.Println("####")
			err = nil
		}
		errCh <- err
	}()

	// 5.
	sigCh = make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	syncErrs := func(count int) {
		for i := 0; i < count; i++ {
			err = errors.Join(err, <-errCh)
		}
	}

	select {
	case e := <-errCh:
		logger.Error("... received from channel errch", "error", e)
		err = errors.Join(err, e)
		count -= 1
	case sig := <-sigCh:
		fmt.Println()
		logger.Info("... received from channel quit", "signal", sig.String())
		err = errors.Join(err, listener.Close())
	}

	syncErrs(count)
}
