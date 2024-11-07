package proxy

import (
	"bytes"
	"fmt"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StdLogger struct {
	*zap.Logger
}

// implements io.Writer for socks5.Config.Logger
func (self *StdLogger) Write(p []byte) (int, error) {
	self.Named("socks5").Log(
		zapcore.ErrorLevel,
		fmt.Sprintf("%s", bytes.TrimSpace(p)),
	)

	return 0, nil
}

func NewStdLogger(lg *zap.Logger) *log.Logger {
	return log.New(&StdLogger{Logger: lg}, "", 0)
}
