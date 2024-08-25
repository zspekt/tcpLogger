package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"

	"gopkg.in/natefinch/lumberjack.v2"
)

func logger() {
	c := setupConfig()

	ch := make(chan []byte, 5)
	defer close(ch) // channels, unlike files, don't need to be closed. doing it anyway

	// lumberjack logger
	logger := c.logger
	defer logger.Close()

	listener, err := net.Listen(c.protocol, c.address+":"+c.port)
	if err != nil {
		slogFatal("error creating listener", "error", err)
	}

	go log(ch, logger)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error accepting connection", "error", err)
		}
		slog.Info("accepted connection without error")

		go handleConn(conn, ch)
	}
}

func handleConn(conn net.Conn, ch chan<- []byte) {
	slog.Info("handling connection...")
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadBytes('\n') // openwrt's logd always sends a newline
		if err != nil {                    // so no need to worry ab this
			switch err {
			case io.EOF:
				slog.Error(
					"EOF while reading from net.Conn reader (conn closed?). killing handleConn goroutine",
				)
				return
			default:
				slog.Error(
					"error reading bytes from conn reader. continuing loop...",
					"error",
					err,
				)
				continue
			}
		}
		ch <- msg
	}
}

func log(ch <-chan []byte, logger *lumberjack.Logger) {
	slog.Info("starting log routine...")
	for msg := range ch {
		_, err := logger.Write(msg)
		if err != nil {
			slog.Error("lumberjack.Logger error writing entry. continuing loop...", "error", err)
			continue
		}
	}
}
