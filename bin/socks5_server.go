package bin

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/d2jvkpn/socks5-proxy/pkg/proxy"

	"github.com/armon/go-socks5"
	"github.com/d2jvkpn/gotk"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func RunSocks5Server(args []string) {
	var (
		err     error
		addr    string
		config  string
		subkey  string
		network string
		fSet    *flag.FlagSet

		logger       *slog.Logger
		zlogger      *gotk.ZapLogger
		conf         *socks5.Config
		handler      *Handler
		listener     net.Listener
		socks5Server *socks5.Server
		errCh        chan error
	)

	fSet = flag.NewFlagSet("socks5_server", flag.ContinueOnError) // flag.ExitOnError

	fSet.StringVar(&addr, "addr", ":1091", "socks5 listening address")
	fSet.StringVar(&config, "config", "", "account authenticator")
	fSet.StringVar(&subkey, "subkey", "accounts", "sub key of accounts in config")
	fSet.StringVar(&network, "network", "tcp", "network")

	fSet.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage of socks5 server:\n")
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
	handler = &Handler{
		Credentials: make(map[string]string),
	}

	conf = &socks5.Config{
		Resolver: handler,
		Logger:   proxy.NewStdLogger(zlogger.Logger),
	}

	if config != "" {
		if handler.Credentials, err = NewCredentials(config, subkey); err != nil {
			return
		}
		conf.AuthMethods = []socks5.Authenticator{
			socks5.UserPassAuthenticator{Credentials: handler},
		}
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
			"Starting SOCKS5 server",
			"command", "server",
			"address", addr,
			"config", config,
			"subkey", subkey,
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

type Handler struct {
	Credentials map[string]string `json:"credentials"`
}

func NewCredentials(fp, key string) (credentials map[string]string, err error) {
	type Account struct {
		Account  string `mapstructure:"account"`
		Password string `mapstructure:"password"`
	}

	var (
		accounts []Account
		vp       *viper.Viper
	)

	accounts = make([]Account, 0)

	vp = viper.New()
	vp.SetConfigType("yaml")
	vp.SetConfigFile(fp)

	if err = vp.UnmarshalKey(key, &accounts); err != nil {
		return nil, err
	}

	credentials = make(map[string]string)

	for i := range accounts {
		credentials[accounts[i].Account] = accounts[i].Password
	}

	return credentials, nil
}

func (self *Handler) Valid(account, password string) (ok bool) {
	var pass string

	pass, ok = self.Credentials[account]

	return ok && password == pass
}

func (self *Handler) Resolve(ctx context.Context, name string) (
	c context.Context, ip net.IP, err error) {

	var ips []net.IP

	if ips, err = net.LookupIP(name); err != nil {
		return
	}

	return ctx, ips[0], err
}
