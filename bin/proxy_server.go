package bin

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/d2jvkpn/socks5-proxy/pkg/proxy"

	"github.com/armon/go-socks5"
	"github.com/d2jvkpn/gotk"
	"go.uber.org/zap"
)

func RunProxyServer(args []string) {
	var (
		err     error
		addr    string
		network string
		fSet    *flag.FlagSet

		logger       *slog.Logger
		zlogger      *gotk.ZapLogger
		conf         *socks5.Config
		listener     net.Listener
		socks5Server *socks5.Server
		errCh        chan error
	)

	fSet = flag.NewFlagSet("server", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&addr, "addr", ":1091", "socks5 listening address")
	fSet.StringVar(&network, "network", "tcp", "network")

	fSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of proxy server:\n")
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

	zlogger, _ = gotk.NewZapLogger("", zap.InfoLevel, 0)
	conf = &socks5.Config{
		Logger: proxy.NewStdLogger(zlogger.Logger),
	}

	if socks5Server, err = socks5.New(conf); err != nil {
		err = fmt.Errorf("Failed to create SOCKS5 server: %w", err)
		return
	}

	if listener, err = net.Listen(network, addr); err != nil {
		err = fmt.Errorf("Failed to bind to %s: %v", addr, err)
		return
	}

	errCh = make(chan error, 1)

	go func() {
		var err error

		logger.Info(
			"Starting SOCKS5 proxying server",
			"command", "server",
			"address", addr,
			"network", network,
		)

		err = socks5Server.Serve(listener)
		// accept tcp [::]:1080: use of closed network connection
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			// fmt.Println("####")
			err = nil
		}
		errCh <- err
	}()

	err = gotk.ExitChan(errCh, listener.Close)
}
