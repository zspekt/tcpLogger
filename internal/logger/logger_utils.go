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
	slog.Info("handling connection...")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// select {
		// case <-shutdwn:
		// 	slog.Info("handleConn received shutdown signal. returning...")
		// 	return
		// default:
		msg, err := ReadBytesWithCtx(reader, '\n', ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Error(
					"EOF while reading from net.Conn reader (conn closed?). returning",
				)
				return
				// continue this doesn't make sense wtf // we continue so we can (probably) handle the shutdown case.
			}
			if errors.Is(err, shutdownErr) {
				slog.Info("handleConn(): caught shutdownErr from ReadBytesWithShutdown()")
				return
				// continue this doesn't make sense wtf // we continue so we can (probably) handle the shutdown case.
			}
			slog.Error( // if we get here, something went wrong :0
				"error reading bytes from conn reader. finishing loop...", "error", err)
		}
		if len(msg) > 0 {
			ch <- msg
		}
		// }
	}
}

func logWithCtx(ch <-chan []byte, logger *lumberjack.Logger, ctx context.Context) {
	slog.Info("starting log routine...")
	for {
		select {
		case <-ctx.Done():
			slog.Info("log(): got cancel signal")
			for i := len(ch); i > 0; i-- {
				logger.Write(<-ch)
			}
			return
		case msg, ok := <-ch:
			slog.Info("log routine got message", "msg", msg)
			if !ok { // channel is closed == we're shutting down (should be last step)
				slog.Info("log routine met with closed channel (shutting down?). returning...")
				return
			}
			_, err := logger.Write(msg)
			if err != nil {
				slog.Error(
					"lumberjack.Logger error writing entry. continuing loop...",
					"error",
					err,
				)
				continue
			}
		}
	}
}

func shutdown(sigs chan os.Signal, cancel context.CancelFunc) {
	// we register the channel so it will get these sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	slog.Info("before sending to shutdown") // we get here
	cancel()
	slog.Info(
		fmt.Sprintf(
			"shutdown routine caught %v sig. cancelling...",
			s.String(),
		),
	) // and after changing the channel buffer size from 0 to 1, we get here
	// however, we never get to the case <- shutdwn
}

func ReadBytesWithCtx(r BytesReader, delim byte, ctx context.Context) ([]byte, error) {
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
			slog.Info("Successfully reading from conn")
			if len(bb) == 0 {
				slog.Error("successfully read MY ASS. this byte slice is empty")
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
