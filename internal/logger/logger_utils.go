package logger

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

type BytesReader interface {
	ReadBytes(byte) ([]byte, error)
}

var shutdownErr error = errors.New("got shutdown signal")

func handleConnWithCtx(conn net.Conn, ch chan<- []byte, ctx context.Context) {
	slog.Info("handleConnWithCtx(): running...")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		slog.Debug("handleConnWithCtx(): running loop...")
		msg, err := ReadBytesWithCtx(reader, '\n', ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Error(
					"handleConnWithCtx(): EOF while reading from net.Conn reader (conn closed?). returning...",
				)
				return
			}
			if errors.Is(err, shutdownErr) {
				slog.Info(
					"handleConnWithCtx(): caught shutdownErr from ReadBytesWithShutdown(). returning...",
				)
				return
			}
			slog.Error(
				"handleConnWithCtx(): error reading bytes from conn reader. finishing loop...",
				"error",
				err,
			)
		}
		if len(msg) > 0 {
			slog.Debug("handleConnWithCtx(): msg not empty. sending to ch...")
			ch <- msg
		}
	}
}

func logWithCtx(ch <-chan []byte, logger *lumberjack.Logger, ctx context.Context) {
	slog.Info("logWithCtx(): starting routine...")
	for {
		select {
		case <-ctx.Done():
			slog.Info("logWithCtx(): got cancel signal. writing to logger and returning...")
			for i := len(ch); i > 0; i-- {
				logger.Write(<-ch)
			}
			return
		case msg, ok := <-ch:
			slog.Debug("logWithCtx(): got message", "msg", msg)
			if !ok { // channel is closed == we're shutting down (should be last step)
				slog.Info("logWithCtx(): channel is closed (shutting down?). returning...")
				return
			}
			_, err := logger.Write(msg)
			if err != nil {
				slog.Error(
					"logWithCtx(): lumberjack.Logger error writing entry. continuing loop...",
					"error",
					err,
				)
				continue
			}
		}
	}
}

func shutdown(sigs chan os.Signal, cancel context.CancelFunc) {
	slog.Info("shutdown(): starting routine...")

	// we register the channel so it will get these sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	slog.Info(
		fmt.Sprintf(
			"shutdown routine caught %v sig. cancelling...",
			s.String(),
		),
	)
	cancel()
}

func ReadBytesWithCtx(r BytesReader, delim byte, ctx context.Context) ([]byte, error) {
	slog.Debug("ReadBytesWithCtx(): called...")
	ch := make(chan struct{})
	var (
		bb  []byte
		err error
	)
	go func() {
		bb, err = r.ReadBytes('\n')
		ch <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			slog.Info("ReadBytesWithCtx(): got cancel signal. returning...")
			return nil, shutdownErr
		case <-ch:
			slog.Debug("ReadBytesWithCtx(): read from conn")
			if len(bb) == 0 {
				slog.Debug("ReadBytesWithCtx(): bytes from conn were empty")
				return nil, err
			}
			return bb, err
		}
	}
}

// AcceptWithCtx is a wrapper around the Accept() method of net.Listener
// that will return nil, shutdownErr if the ctx i canceled before a connection
// is accepted. Otherwise, returns conn, nil
func AcceptWithCtx(l net.Listener, ctx context.Context) (net.Conn, error) {
	slog.Info("AcceptWithCtx(): called...")
	var (
		conn net.Conn
		err  error
	)

	ch := make(chan struct{}, 1)
	go func() {
		conn, err = l.Accept()
		ch <- struct{}{}
	}()
	for {
		select {
		case <-ch:
			if err != nil {
				return nil, err
			}
			return conn, nil
		case <-ctx.Done():
			slog.Info("AcceptWithTimeout(): caught cancel signal while waiting for conn")
			return nil, shutdownErr
		}
	}
}
