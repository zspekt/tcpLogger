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
	slog.Info("logger.Run(): running...")
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	go shutdown(sigs, cancel)

	logger := c.Logger
	defer logger.Close()
	ch := make(chan []byte, 5)
	go logWithCtx(ch, logger, ctx)

	listener, err := net.Listen(c.Protocol, c.Address+":"+c.Port)
	utils.Must(err)

	defer listener.Close()

	// declaring it here so we can close it on case <-shutdwn
	var conn net.Conn
	for {
		slog.Debug("logger.Run(): running main loop...")
		select {
		case <-ctx.Done():
			slog.Info("logger.Run(): received cancel sig...")
			if conn != nil {
				slog.Info("logger.Run(): closing current connection...")
				err := conn.Close()
				if err != nil {
					slog.Error("logger.Run(): error closing connection", "error", err)
				}
			}
			slog.Info("logger.Run(): closing channel...")
			close(ch)

			slog.Info("logger.Run(): closing logger...")
			err = logger.Close()
			if err != nil {
				slog.Error("logger.Run(): error closing logger", "error", err)
			}
			return
		default:
			slog.Debug("logger.Run(): running default case on select loop...")
			conn, err = AcceptWithCtx(listener, ctx)
			if err != nil {
				slog.Error("logger.Run(): error accepting connection", "error", err)
				continue
			}
			slog.Debug("logger.Run(): accepted connection without error")
			handleConnWithCtx(conn, ch, ctx)
		}
	}
}
