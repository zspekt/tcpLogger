package utils

import (
	"log/slog"
	"os"
)

// slog equivalent of log.Fatal()
func SlogFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func Must(err error) {
	if err != nil {
		SlogFatal(err.Error())
	}
}
