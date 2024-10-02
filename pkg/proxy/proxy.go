package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

type Proxy struct {
	SSH_Address    string `mapstructure:"ssh_address"`
	SSH_User       string `mapstructure:"ssh_user"`
	SSH_Password   string `mapstructure:"ssh_password"`
	SSH_PrivateKey string `mapstructure:"ssh_private_key"`
	SSH_KnownHosts string `mapstructure:"ssh_known_hosts"`

	Socks5User     string `mapstructure:"socks5_user"`
	Socks5Password string `mapstructure:"socks5_password"`

	*ssh.Client `mapstructure:"-"`
	Logger      *Logger `mapstructure:"-"`
}

type Logger struct {
	*zap.Logger
}

func NewLogger(lg *zap.Logger) *Logger {
	return &Logger{Logger: lg}
}

// implements io.Writer for socks5.Config.Logger
func (self *Logger) Write(p []byte) (int, error) {
	self.Named("socks5").Log(
		zapcore.ErrorLevel,
		fmt.Sprintf("%s", bytes.TrimSpace(p)),
	)
	return 0, nil
}

func (self *Logger) StdLogger() *log.Logger {
	return log.New(self, "", 0)
}

func LoadProxy(fp string, key string, logger *zap.Logger) (config *Proxy, err error) {
	var vp *viper.Viper

	vp = viper.New()
	vp.SetConfigType("yaml")

	// vp.SetConfigName("config")
	vp.SetConfigFile(fp)

	if err = vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	config = new(Proxy)
	err = vp.UnmarshalKey(key, config)
	// err = vp.Unmarshal(config)

	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if config.SSH_Address == "" || config.SSH_User == "" {
		return nil, fmt.Errorf("ssh_address or ssh_user is empty")
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
		config.Logger = &Logger{Logger: lg.Logger}
	} else {
		config.Logger = &Logger{Logger: logger}
	}

	return config, nil
}

func (self *Proxy) dial() (err error) {
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

	if self.SSH_KnownHosts != "" {
		if config.HostKeyCallback, err = knownhosts.New(self.SSH_KnownHosts); err != nil {
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

	if self.Client, err = ssh.Dial("tcp", self.SSH_Address, config); err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	return nil
}

// ssh auth methods
func (self *Proxy) AuthMethods() string {
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

func (self *Proxy) Resolve(ctx context.Context, name string) (
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
			logger.Error("resolve", zap.String("name", name), zap.Any("error", &err))
		} else {
			logger.Debug("resolve", zap.String("name", name))
		}
	}()

	if session, err = self.Client.NewSession(); err != nil {
		err = fmt.Errorf("unable to create ssh session: %w", err)
		return
	}
	defer session.Close()

	if reader, err = session.StdoutPipe(); err != nil {
		err = fmt.Errorf("unable to create stdout pipe: %w", err)
		return
	}

	if err = session.Start(fmt.Sprintf("dig +short %s", name)); err != nil {
		err = fmt.Errorf("unable to run dig +short: %w", err)
		return
	}

	bts, err = io.ReadAll(reader)
	ip = net.ParseIP(string(bts))

	return
}

func (self *Proxy) Socks5Config() (config *socks5.Config) {
	var logger *zap.Logger

	logger = self.Logger.Named("proxy")

	config = &socks5.Config{
		Resolver: self, // socks5.DNSResolver{},
		Logger:   self.Logger.StdLogger(),
		Dial: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			// println("~~~", network, addr)
			conn, err = self.Client.Dial(network, addr)

			if err != nil {
				logger.Warn(
					"ssh dail",
					zap.String("network", network),
					zap.String("addr", addr),
					zap.Any("error", err),
				)
			} else {
				logger.Debug(
					"ssh dail",
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

func (self *Proxy) Close() (err error) {
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
