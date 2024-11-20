package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"

	"github.com/armon/go-socks5"
	"github.com/d2jvkpn/gotk"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Socks5SSH struct {
	SSH_Host       string `mapstructure:"ssh_host"`
	SSH_Port       int    `mapstructure:"ssh_port"`
	SSH_User       string `mapstructure:"ssh_user"`
	SSH_Password   string `mapstructure:"ssh_password"`
	SSH_PrivateKey string `mapstructure:"ssh_private_key"`
	SSH_KnownHosts string `mapstructure:"ssh_known_hosts"`

	Socks5User     string `mapstructure:"socks5_user"`
	Socks5Password string `mapstructure:"socks5_password"`

	Viper       *viper.Viper `mapstructure:"-"`
	Logger      *zap.Logger  `mapstructure:"-"`
	*ssh.Client `mapstructure:"-"`
}

func LoadProxy(fp string, key string, logger *zap.Logger) (
	config *Socks5SSH, err error) {
	var vp *viper.Viper

	vp = viper.New()
	vp.SetConfigType("yaml")

	// vp.SetConfigName("config")
	vp.SetConfigFile(fp)

	if err = vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	config = new(Socks5SSH)
	config.Viper = vp

	// err = vp.Unmarshal(config)
	if err = vp.UnmarshalKey(key, config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	vp.SetDefault("ssh_port", 22)

	if config.SSH_Host == "" || config.SSH_User == "" {
		return nil, fmt.Errorf("ssh_host or ssh_user is empty")
	}

	switch {
	case config.SSH_Password != "":
	case config.SSH_PrivateKey != "":
	default:
		return nil, fmt.Errorf("no ssh auth")
	}

	if err = config.dial(); err != nil {
		return nil, fmt.Errorf("dial ssh: %w", err)
	}

	if logger == nil {
		lg, _ := gotk.NewZapLogger("", zapcore.InfoLevel, 0)
		config.Logger = lg.Logger
	} else {
		config.Logger = logger
	}

	return config, nil
}

func (self *Socks5SSH) dial() (err error) {
	var (
		config *ssh.ClientConfig
		signer ssh.Signer
		auths  []ssh.AuthMethod
	)

	defer func() {
		if err == nil {
			return
		}

		err = errors.Join(err, self.Close())
	}()

	config = &ssh.ClientConfig{User: self.SSH_User}

	// TODO: ssh-keyscan -p port host
	if self.SSH_KnownHosts != "" {
		config.HostKeyCallback, err = knownhosts.New(self.SSH_KnownHosts)
		if err != nil {
			return fmt.Errorf("loading known hosts: %w", err)
		}
	} else {
		// Warning: use this only for testing, not in production
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	auths = make([]ssh.AuthMethod, 0)
	if self.SSH_Password != "" {
		auths = append(auths, ssh.Password(self.SSH_Password))
	}

	if self.SSH_PrivateKey != "" {
		if signer, err = LoadPrivateKey(self.SSH_PrivateKey); err != nil {
			return fmt.Errorf("loading private key: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	config.Auth = auths

	self.Client, err = ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", self.SSH_Host, self.SSH_Port),
		config,
	)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	return nil
}

// ssh auth methods
func (self *Socks5SSH) AuthMethods() string {
	var methods []string

	methods = make([]string, 0)

	if self.SSH_Password != "" {
		methods = append(methods, "password")
	}

	if self.SSH_PrivateKey != "" {
		methods = append(methods, "private_key")
	}

	return strings.Join(methods, ",")
}

func (self *Socks5SSH) Resolve(ctx context.Context, name string) (
	c context.Context, ip net.IP, err error) {

	var (
		bts     []byte
		session *ssh.Session
		reader  io.Reader
		logger  *zap.Logger
	)

	logger = self.Logger.Named("proxy")

	defer func() {
		if err != nil {
			logger.Error("resolve", zap.String("name", name), zap.Any("error", err))
		} else {
			logger.Debug("resolve", zap.String("name", name))
		}
	}()

	if session, err = self.Client.NewSession(); err != nil {
		err = fmt.Errorf("unable to create ssh session: %w", err)
		return ctx, ip, err
	}
	defer session.Close()

	if reader, err = session.StdoutPipe(); err != nil {
		err = fmt.Errorf("unable to create stdout pipe: %w", err)
		return ctx, ip, err
	}

	if err = session.Start(fmt.Sprintf("dig +short %s", name)); err != nil {
		err = fmt.Errorf("unable to run dig +short: %s, %w", name, err)
		return ctx, ip, err
	}

	if bts, err = io.ReadAll(reader); err != nil {
		return ctx, ip, err
	}

	ip = net.ParseIP(string(bts))

	return ctx, ip, err
}

func (self *Socks5SSH) Socks5Config() (config *socks5.Config) {
	var logger *zap.Logger

	logger = self.Logger.Named("proxy")

	config = &socks5.Config{
		Resolver: self, // default: socks5.DNSResolver{},
		Logger:   NewStdLogger(self.Logger),
		Dial: func(ctx context.Context, network, addr string) (
			conn net.Conn, err error) {
			// println("~~~", network, addr)
			conn, err = self.Client.Dial(network, addr)

			if err != nil {
				logger.Warn(
					"dial",
					zap.String("network", network),
					zap.String("addr", addr),
					zap.Any("error", err),
				)
			} else {
				logger.Debug(
					"dial",
					zap.String("network", network),
					zap.String("addr", addr),
				)
			}

			return conn, err
		},
	}

	if self.Socks5User != "" {
		credentials := socks5.StaticCredentials{
			self.Socks5User: self.Socks5Password,
		}

		config.AuthMethods = []socks5.Authenticator{
			socks5.UserPassAuthenticator{Credentials: credentials},
		}
	}

	return config
}

func (self *Socks5SSH) Close() (err error) {
	if self == nil {
		return
	}

	if self.Client != nil {
		err = errors.Join(err, self.Client.Close())
		self.Client = nil
	}

	//if self.Logger != nil {
	//	err = errors.Join(err, self.Logger.Sync())
	//}

	return err
}

// path: /home/account/.ssh/id_rsa
// LoadPrivateKey loads an RSA private key from a file
func LoadPrivateKey(p string) (signer ssh.Signer, err error) {
	var bts []byte

	// Read the private key file
	if bts, err = ioutil.ReadFile(p); err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	// Parse the private key
	if signer, err = ssh.ParsePrivateKey(bts); err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return signer, nil
}

// ?? path: /home/account/.ssh/id_rsa.pub
func LoadAuthorizedKey(p string) (pubKey ssh.PublicKey, err error) {
	var bts []byte

	if bts, err = ioutil.ReadFile(p); err != nil {
		return nil, fmt.Errorf("unable to read public key: %w", err)
	}

	// ssh.ParseKnownHosts
	if pubKey, _, _, _, err = ssh.ParseAuthorizedKey(bts); err != nil {
		return nil, fmt.Errorf("unable to parse authorized key: %w", err)
	}

	if pubKey, err = ssh.ParsePublicKey(pubKey.Marshal()); err != nil {
		return nil, fmt.Errorf("unable to parse public key: %w", err)
	}

	return pubKey, nil
}

// ?? path: /home/account/.ssh/known_hosts
func LoadKnownHosts(keyPath string) (pubKey ssh.PublicKey, err error) {
	var bts []byte

	if bts, err = ioutil.ReadFile(keyPath); err != nil {
		return nil, fmt.Errorf("unable to read public key: %w", err)
	}

	if _, _, pubKey, _, _, err = ssh.ParseKnownHosts(bts); err != nil {
		return nil, fmt.Errorf("unable to parse authorized key: %w", err)
	}

	if pubKey, err = ssh.ParsePublicKey(pubKey.Marshal()); err != nil {
		return nil, fmt.Errorf("unable to parse public key: %w", err)
	}

	return pubKey, nil
}
