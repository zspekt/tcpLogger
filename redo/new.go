package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func slogFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func getEnvOrDefault(key string, def string) (string, error) {
	switch {
	case key == "" && def == "":
		return "", errors.New("caller passed empty strings for key and def arg")
	case key == "":
		return "", errors.New("caller passed an empty string as key arg")
	case def == "":
		return "", errors.New("caller passed an empty string as def arg")
	}

	// os.LookupEnv returns false if var is not set
	v, ok := os.LookupEnv(key)
	switch {
	case !ok:
		slog.Info(fmt.Sprintf("env var <%v> not set up. using default <%v>", key, def))
		return def, nil
	case v == "":
		slog.Info(fmt.Sprintf("env var <%v> is set up, but empty. using default <%v>", key, def))
		return def, nil
	}
	return v, nil
}

func setup() {
	filename, err := getEnvOrDefault("FILENAME", "/var/log/openwrt/openwrt.log")
	if err != nil {
		slogFatal(err.Error())
	}

	port, err := getEnvOrDefault("PORT", "8080")
	if err != nil {
		slogFatal(err.Error())
	}

	protocol, err := getEnvOrDefault("PROTOCOL", "tcp")
	if err != nil {
		slogFatal(err.Error())
	}

	address, err := getEnvOrDefault("ADDRESS", "localhost")

	print(filename, port, protocol, address)
}
