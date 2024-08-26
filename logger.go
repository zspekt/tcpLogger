// TODO: handleConn, concurrent call or not?
//       if yes, channel to communicate shutdown? flush writer?

//			OR:
//		      if  no, SIGINT/TERM would have to go into handleConn,
//		      which could return a specific error type.
//		      the type of the error would change the flow of the loop,
//	        making it terminate
package main

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

func logger() {
	c := setupConfig()

	ch := make(chan []byte, 5)
	logger := c.logger
	defer logger.Close()
	go log(ch, logger)

	sigs := make(chan os.Signal, 1)
	shutdwn := make(chan os.Signal, 1)
	go shutdown(shutdwn, sigs)

	listener, err := net.Listen(c.protocol, c.address+":"+c.port)
	if err != nil {
		slogFatal("error creating listener", "error", err)
	}

	// declaring it here so we can close it on case <-shutdwn

	var conn net.Conn
	for {
		slog.Debug("running main for loop...")
		select {
		case s := <-shutdwn:
			slog.Info(fmt.Sprintf("logger received %v signal...", s))
			if conn != nil {
				slog.Info("closing current connection...")
				err := conn.Close()
				if err != nil {
					slog.Error("error closing connection", "error", err)
				}
			}
			time.Sleep(5 * time.Second)
			slog.Info("closing channel...")
			close(ch)

			slog.Info("closing logger...")
			err = logger.Close()
			if err != nil {
				slog.Error("error closing logger", "error", err)
			}
			return
		default:
			slog.Debug("running default case on main for select loop...")
			conn, err = AcceptWithTimeout(listener, 3*time.Second)
			if err != nil {
				slog.Error("error accepting connection", "error", err)
				continue
			}
			slog.Info("accepted connection without error")
			handleConn(conn, ch, shutdwn)
		}
	}
}

func handleConn(conn net.Conn, ch chan<- []byte, shutdwn <-chan os.Signal) {
	slog.Info("handling connection...")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	reader.WriteTo(&lumberjack.Logger{})
	for {
		select {
		case <-shutdwn:

		default:
			msg, err := reader.ReadBytes('\n') // openwrt's logd always sends a newline
			if err != nil {                    // so no need to worry ab this
				if err == io.EOF {
					slog.Error(
						"EOF while reading from net.Conn reader (conn closed?). continuing loop...",
					)
					return
				} else {
					slog.Error(
						"error reading bytes from conn reader. finishing loop...",
						"error",
						err,
					)
				}
			}
			ch <- msg
		}
	}
}

func log(ch <-chan []byte, logger *lumberjack.Logger) {
	slog.Info("starting log routine...")
	for {
		msg, ok := <-ch
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

func shutdown(shutdown chan os.Signal, sigs chan os.Signal) {
	// we register the channel so it will get these sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	slog.Info("before sending to shutdown") // we get here
	shutdown <- s
	slog.Info(
		fmt.Sprintf("shutdown routine caught %v sig. notifying logger routine...", s.String()),
	) // and after changing the channel buffer size from 0 to 1, we get here
	// however, we never get to the case <- shutdwn
}

// AcceptWithTimeout is a wrapper around the Accept() method of net.Listener
// that will return a net.OpError if the timeout is reached before a connection
// is accepted. Otherwise, returns conn, nil
func AcceptWithTimeout(l net.Listener, d time.Duration) (net.Conn, error) {
	t := time.After(d)
	ch := make(chan struct{})

	var (
		conn net.Conn
		err  error
	)
	go func() {
		conn, err = l.Accept()
		ch <- struct{}{}
	}()
	for {
		select {
		case <-t:
			return nil, &net.OpError{
				Op:     "accept",
				Net:    "tcp",
				Source: nil,
				Addr:   nil,
				Err:    errors.New("listener accept conn timeout"),
			}
		case <-ch:
			if err != nil {
				return nil, err
			}
			return conn, nil
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
