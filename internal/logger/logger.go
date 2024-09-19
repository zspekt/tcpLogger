// TODO: replace 46 gazillion channels with a single ctx ?
package logger

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/zspekt/tcpLogger/internal/setup"
	"github.com/zspekt/tcpLogger/internal/utils"
)

func Run(c *setup.Cfg) {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	go shutdownWithCtx(sigs, cancel)

	logger := c.Logger
	defer logger.Close()
	ch := make(chan []byte, 5)
	go logWithCtx(ch, logger, ctx)

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
		case <-ctx.Done():
			slog.Info("RunWithCtx(): received cancel sig...")
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
			conn, err = AcceptWithCtx(listener, ctx)
			if err != nil {
				slog.Error("logger(): error accepting connection", "error", err)
				continue
			}
			slog.Info("logger(): accepted connection without error")
			handleConnWithCtx(conn, ch, ctx)
		}
	}
}
