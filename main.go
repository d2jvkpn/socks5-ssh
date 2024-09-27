package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net"
	"os"

	"github.com/armon/go-socks5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
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
		proxyConfig  *ProxyConfig
		sshClient    *ssh.Client
		socks5Config *socks5.Config
		listener     net.Listener
		socks5Server *socks5.Server
	)

	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

	flag.StringVar(&config, "config", "configs/local.yaml", "configuration file(yaml)")
	flag.StringVar(&subkey, "subkey", "socks5_ssh", "use subkey of config(yaml)")
	flag.StringVar(&addr, "addr", ":1080", "socks5 listening address")
	flag.StringVar(&network, "network", "tcp", "network")
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.Parse()

	defer func() {
		if sshClient != nil {
			_ = sshClient.Close()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}()

	if proxyConfig, err = LoadProxyConfig(config, subkey); err != nil {
		err = fmt.Errorf("Failed to load ssh config: %w", err)
		return
	}

	if sshClient, err = proxyConfig.DialSSH(); err != nil {
		err = fmt.Errorf("Failed to dial ssh: %w", err)
		return
	}

	socks5Config = &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			// println("~~~", network, addr)
			conn, err = sshClient.Dial(network, addr)

			if err != nil {
				if debug {
					logger.Warn("ssh dail", "network", network, "addr", addr, "error", err)
				}
				return nil, err
			}

			if debug {
				logger.Info("ssh dail", "network", network, "addr", addr)
			}

			return conn, nil
		},
	}

	if proxyConfig.Socks5User != "" {
		credentials := socks5.StaticCredentials{
			proxyConfig.Socks5User: proxyConfig.Socks5Password,
		}

		// fmt.Println("~~~", credentials)
		socks5Config.AuthMethods = []socks5.Authenticator{
			socks5.UserPassAuthenticator{Credentials: credentials},
		}
	}

	if socks5Server, err = socks5.New(socks5Config); err != nil {
		err = fmt.Errorf("Failed to create SOCKS5 server: %w", err)
		return
	}

	if listener, err = net.Listen(network, addr); err != nil {
		err = fmt.Errorf("Failed to dail %s: %w", addr, err)
		return
	}

	logger.Info(
		"Starting SOCKS5 proxy",
		"config", config,
		"address", addr,
		"debug", debug,
		"network", network,
		"auth", proxyConfig.Socks5User != "",
	)

	if err = socks5Server.Serve(listener); err != nil {
		err = fmt.Errorf("Failed to start SOCKS5 server: %w", err)
		return
	}
}

type ProxyConfig struct {
	SSH_Address    string `mapstructure:"ssh_address"`
	SSH_User       string `mapstructure:"ssh_user"`
	SSH_Password   string `mapstructure:"ssh_password"`
	SSH_PrivateKey string `mapstructure:"ssh_private_key"`
	SSH_PublicKey  string `mapstructure:"ssh_public_key"`
	Socks5User     string `mapstructure:"socks5_user"`
	Socks5Password string `mapstructure:"socks5_password"`
}

func LoadProxyConfig(fp string, keys ...string) (config *ProxyConfig, err error) {
	var vp *viper.Viper

	vp = viper.New()
	vp.SetConfigType("yaml")

	// vp.SetConfigName("config")
	vp.SetConfigFile(fp)

	if err = vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	config = new(ProxyConfig)
	if len(keys) > 0 && keys[0] != "" {
		err = vp.UnmarshalKey(keys[0], config)
	} else {
		err = vp.Unmarshal(config)
	}
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if config.SSH_Address == "" || config.SSH_User == "" {
		return nil, fmt.Errorf("ssh_address or ssh_user is empty")
	}

	switch {
	case config.SSH_Password != "":
	case config.SSH_PrivateKey != "" && config.SSH_PublicKey != "":
	default:
		return nil, fmt.Errorf("no ssh auth")
	}

	return config, nil
}

func (self *ProxyConfig) DialSSH() (client *ssh.Client, err error) {
	var (
		config  *ssh.ClientConfig
		hostKey ssh.PublicKey
		signer  ssh.Signer
	)

	config = &ssh.ClientConfig{
		User: self.SSH_User,
	}

	if self.SSH_PublicKey == "" {
		// Warning: use this only for testing, not in production
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else {
		if hostKey, err = LoadPublicKey(self.SSH_PublicKey); err != nil {
			return nil, fmt.Errorf("loading public key: %w", err)
		}
		config.HostKeyCallback = ssh.FixedHostKey(hostKey)
	}

	if self.SSH_Password != "" {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(self.SSH_Password),
		}
	} else {
		if signer, err = LoadPrivateKey(self.SSH_PrivateKey); err != nil {
			return nil, fmt.Errorf("loading private key: %w", err)
		}

		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

	if client, err = ssh.Dial("tcp", self.SSH_Address, config); err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	return client, nil
}

// path: /home/account/.ssh/id_rsa
// LoadPrivateKey loads an RSA private key from a file
func LoadPrivateKey(keyPath string) (signer ssh.Signer, err error) {
	var bts []byte

	// Read the private key file
	if bts, err = ioutil.ReadFile(keyPath); err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	// Parse the private key
	if signer, err = ssh.ParsePrivateKey(bts); err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return signer, nil
}

// path: /home/account/.ssh/id_rsa.pub
func LoadPublicKey(keyPath string) (pubKey ssh.PublicKey, err error) {
	var (
		bts []byte
		pk  ssh.PublicKey
	)

	if bts, err = ioutil.ReadFile(keyPath); err != nil {
		return nil, fmt.Errorf("unable to read public key: %w", err)
	}

	if pk, _, _, _, err = ssh.ParseAuthorizedKey(bts); err != nil {
		return nil, fmt.Errorf("unable to parse authorized key: %w", err)
	}

	if pubKey, err = ssh.ParsePublicKey(pk.Marshal()); err != nil {
		return nil, fmt.Errorf("unable to parse public key: %w", err)
	}

	return pubKey, nil
}
