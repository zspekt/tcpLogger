package logger

import (
	"net"
	"strings"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/zspekt/tcpLogger/internal/setup"
)

func Test_logger(t *testing.T) {
	tests := []struct {
		name string
		arg  *setup.Cfg
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

			c := &setup.Cfg{
				Port:     addr[1],
				Logger:   &lumberjack.Logger{},
				Protocol: "tcp",
				Address:  addr[0],
			}

			Run(c)
		})
	}
}
