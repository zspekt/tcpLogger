package logger

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type BytesReader interface {
	ReadBytes(byte) ([]byte, error)
}

var shutdownErr error = errors.New("got shutdown signal")

func handleConn(conn net.Conn, ch chan<- []byte, shutdwn <-chan struct{}) {
	slog.Info("handling connection...")
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		// select {
		// case <-shutdwn:
		// 	slog.Info("handleConn received shutdown signal. returning...")
		// 	return
		// default:
		msg, err := ReadBytesWithShutdown(reader, '\n', shutdwn)
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
	}
}

func log(ch <-chan []byte, logger *lumberjack.Logger) {
	slog.Info("starting log routine...")
	for {
		msg, ok := <-ch
		slog.Info("log routine got message", "msg", msg)
		if !ok { // channel is closed == we're shutting down (should be last step)
			slog.Info("log routine met with closed channel (shutting down?). returning...")
			return
		}
		_, err := logger.Write(msg)
		if err != nil {
			slog.Error("lumberjack.Logger error writing entry. continuing loop...", "error", err)
			continue
		}
	}
}

func shutdown(
	stopLoggerLoop, stopHandleConn chan struct{},
	stopAcceptConn chan struct{},
	sigs chan os.Signal,
) {
	// we register the channel so it will get these sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	slog.Info("before sending to shutdown") // we get here
	stopLoggerLoop <- struct{}{}
	stopAcceptConn <- struct{}{}
	stopHandleConn <- struct{}{}
	slog.Info(
		fmt.Sprintf(
			"shutdown routine caught %v sig. notifying logger routine...",
			s.String(),
		),
	) // and after changing the channel buffer size from 0 to 1, we get here
	// however, we never get to the case <- shutdwn
}

func ReadBytesWithShutdown(r BytesReader, delim byte, shutdown <-chan struct{}) ([]byte, error) {
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
		case <-shutdown:
			slog.Info("ReadBytesWithShutdown(): got shutdown signal. returning...")
			return nil, errors.New("got shutdown signal")
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

func ReadBytesWithTimeout(r BytesReader, delim byte, d time.Duration) ([]byte, error) {
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
		case <-time.After(d):
			slog.Error("timeout reading from conn. returning ReadTimeoutError")
			return nil, ReadTimeoutError
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

// AcceptWithTimeout is a wrapper around the Accept() method of net.Listener
// that will return a net.OpError if the timeout is reached before a connection
// is accepted. Otherwise, returns conn, nil
func AcceptWithTimeout(l net.Listener, d time.Duration, shutdwn chan struct{}) (net.Conn, error) {
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
		case <-time.After(d):
			return nil, NetTimeoutError
		case <-ch:
			if err != nil {
				return nil, err
			}
			return conn, nil
		case <-shutdwn:
			shutdwn <- struct{}{}
			slog.Info("AcceptWithTimeout(): caught shutdown signal while waiting for conn")
			return nil, errors.New("shutdown sig while waiting for conn")
		}
	}
}

func AcceptWithShutdown(l net.Listener, shutdwn chan struct{}) (net.Conn, error) {
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
		case <-shutdwn:
			shutdwn <- struct{}{}
			slog.Info("AcceptWithTimeout(): caught shutdown signal while waiting for conn")
			return nil, shutdownErr
		}
	}
}

// func connLoop(listener net.Listener, ch chan<- []byte, shutdown bool) {
// 	for !shutdown {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			slog.Error("error accepting connection", "error", err)
// 		}
// 		slog.Info("accepted connection without error")
//
// 		go handleConn(conn, ch)
// 	} // if we get here, we caught a SIGINT/SIGTERM
// }

// func connLoop(listener net.Listener, ch chan<- []byte, sigChan chan os.Signal) {
// 	for {
// 		select {
// 		case msg := <-ch:
// 			conn, err := listener.Accept()
// 			if err != nil {
// 				slog.Error("error accepting connection", "error", err)
// 			}
// 			slog.Info("accepted connection without error")
//
// 			go handleConn(conn, ch)
// 		case sig := <-sigChan:
// 			break
// 		}
// 	}
// }
