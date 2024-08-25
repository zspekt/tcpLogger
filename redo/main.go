package main

import (
	"bufio"
	"log/slog"
	"net"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	logger()
}

func logger() {
	c := setupConfig()

	logger := c.logger
	defer logger.Close()

	listener, err := net.Listen(c.protocol, c.address+":"+c.port)
	if err != nil {
		slogFatal("error creating listener", "error", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Info("error accepting connection. this will be logged", "error", err)
			logger.Write([]byte(err.Error()))
		}

		go handleConn(conn, logger)
	}
}

func handleConn(conn net.Conn, logger *lumberjack.Logger) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadBytes('\n') // openwrt's logd always sends a newline
		if err != nil {                    // so no need to worry ab this

			slog.Info("error reading bytes from conn reader. this will be logged", "error", err)
			logger.Write([]byte(err.Error()))

			continue
		}

		logger.Write(msg)
	}
}
