package main

import (
	"net"
	"strings"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"
)

func Test_logger(t *testing.T) {
	tests := []struct {
		name string
		arg  *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatal(err)
			}
			addr := strings.Split(l.Addr().String(), ":")
			l.Close()

			c := &Config{
				port:     addr[1],
				logger:   &lumberjack.Logger{},
				protocol: "tcp",
				address:  addr[0],
			}

			logger(c)
		})
	}
}
