package proxy

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// ticker = time.NewTicker(5 * time.Second)
func ProxyHealthCmd(proxy *Socks5SSH, ticker *time.Ticker, maxRestries int) (
	err error) {
	var (
		ok     bool
		count  int
		logger *zap.Logger
	)

	logger = proxy.Logger.Named("proxy")

	healthCheck := func() (err error) {
		var session *ssh.Session

		if session, err = proxy.Client.NewSession(); err != nil {
			err = fmt.Errorf("unable to create ssh session: %w", err)
			return
		}
		defer session.Close()

		if _, err = session.StdoutPipe(); err != nil {
			err = fmt.Errorf("unable to create stdout pipe: %w", err)
			return
		}

		if err = session.Start("echo -n"); err != nil {
			err = fmt.Errorf("unable to run 'echo -n'")
			return
		}

		return
	}

	healthCheckN := func() (err error) {
		for i := 0; i < maxRestries; i++ {
			if err = healthCheck(); err == nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}

		return err
	}

	for {
		select {
		case _, ok = <-ticker.C:
			if !ok {
				logger.Debug("ticker stopped")
				return nil
			}

			if err = healthCheckN(); err != nil {
				logger.Error("health_check", zap.Int("count", count), zap.Any("error", err))
				return err
			} else {
				logger.Debug("health")
			}
		}
	}
}
