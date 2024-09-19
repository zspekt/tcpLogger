// TODO: replace 46 gazillion channels with a single ctx ?
package logger

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/zspekt/tcpLogger/internal/setup"
	"github.com/zspekt/tcpLogger/internal/utils"
)

func Run(c *setup.Cfg) {
	ch := make(chan []byte, 5)
	logger := c.Logger
	defer logger.Close()
	go log(ch, logger)

	sigs := make(chan os.Signal, 1)
	stopAcceptConn := make(chan struct{}, 1)
	stopLoggerRun := make(chan struct{}, 1)
	stopHandleConn := make(chan struct{}, 1)
	go shutdown(stopLoggerRun, stopHandleConn, stopAcceptConn, sigs)

	listener, err := net.Listen(c.Protocol, c.Address+":"+c.Port)
	if err != nil {
		utils.SlogFatal("logger(): fatal error creating listener", "error", err)
	}
	defer listener.Close()

	// declaring it here so we can close it on case <-shutdwn
	var conn net.Conn
	for {
		slog.Info("logger(): running main for loop...")
		select {
		case s := <-stopLoggerRun:
			slog.Info(fmt.Sprintf("logger(): received %v signal...", s))
			if conn != nil {
				slog.Info("logger(): closing current connection...")
				err := conn.Close()
				if err != nil {
					slog.Error("logger(): error closing connection", "error", err)
				}
			}
			slog.Info("logger(): closing channel...")
			close(ch)

			slog.Info("logger(): closing logger...")
			err = logger.Close()
			if err != nil {
				slog.Error("logger(): error closing logger", "error", err)
			}
			return
		default:
			slog.Info("logger(): running default case on main for select loop...")
			// replace with AcceptWithShutdown() probably
			// conn, err = AcceptWithTimeout(listener, 60*time.Second, shutdwn)
			conn, err = AcceptWithShutdown(listener, stopAcceptConn)
			if err != nil {
				slog.Error("logger(): error accepting connection", "error", err)
				continue
			}
			slog.Info("logger(): accepted connection without error")
			handleConn(conn, ch, stopHandleConn)
		}
	}
}
