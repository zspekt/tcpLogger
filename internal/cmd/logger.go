package cmd

import (
	"github.com/zspekt/tcpLogger/internal/logger"
	"github.com/zspekt/tcpLogger/internal/setup"
)

func Run() {
	c := setup.Config()
	logger.Run(c)
}
