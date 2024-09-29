package proxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net"
	"strings"

	"github.com/armon/go-socks5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Proxy struct {
	SSH_Address    string `mapstructure:"ssh_address"`
	SSH_User       string `mapstructure:"ssh_user"`
	SSH_Password   string `mapstructure:"ssh_password"`
	SSH_PrivateKey string `mapstructure:"ssh_private_key"`
	SSH_KnownHosts string `mapstructure:"ssh_known_hosts"`
	sshClient      *ssh.Client

	Socks5User     string `mapstructure:"socks5_user"`
	Socks5Password string `mapstructure:"socks5_password"`
}

func LoadProxy(fp string, keys ...string) (config *Proxy, err error) {
	var vp *viper.Viper

	vp = viper.New()
	vp.SetConfigType("yaml")

	// vp.SetConfigName("config")
	vp.SetConfigFile(fp)

	if err = vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	config = new(Proxy)
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
	case config.SSH_PrivateKey != "":
	default:
		return nil, fmt.Errorf("no ssh auth")
	}

	if err = config.dial(); err != nil {
		return nil, fmt.Errorf("dial ssh: %w", err)
	}

	return config, nil
}

func (self *Proxy) dial() (err error) {
	var (
		config *ssh.ClientConfig
		signer ssh.Signer
		auths  []ssh.AuthMethod
	)

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

	if self.sshClient, err = ssh.Dial("tcp", self.SSH_Address, config); err != nil {
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

func (self *Proxy) Socks5Config(logger *slog.Logger) (config *socks5.Config) {
	config = &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
			// println("~~~", network, addr)
			conn, err = self.sshClient.Dial(network, addr)

			if err != nil {
				logger.Warn("ssh dail", "network", network, "addr", addr, "error", err)
			} else {
				logger.Info("ssh dail", "network", network, "addr", addr)
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
	if self == nil || self.sshClient == nil {
		return nil
	}

	return self.sshClient.Close()
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
