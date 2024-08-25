package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"
)

func getEnvOrDefaultString(key, def string) (string, error) {
	v, err := getEnvOrDefaultGen(key, def)
	if err != nil {
		return "", err
	}
	return v, nil
}

func getEnvOrDefaultInt(key string, def int) (int, error) {
	v, err := getEnvOrDefaultGen(key, def)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func getEnvOrDefaultBool(key string, def bool) (bool, error) {
	v, err := getEnvOrDefaultGen(key, def)
	if err != nil {
		return false, err
	}
	return v, nil
}

func getEnvOrDefaultGen[T string | bool | int](key string, def T) (T, error) {
	var v T

	// if this is a string..
	if defStr, ok := any(def).(string); ok {
		switch {

		// we can only check for zero values on strings
		// as both a 0 int, and a false bool aren't
		// invalid in this situation

		// err: "",
		case key == "" && defStr == "":
			return v, &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"key", "def"},
			}
		case key == "":
			return v, &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"key"},
			}
		case defStr == "":
			return v, &ArgError{
				Err:   "empty string passed as argument",
				Param: []string{"def"},
			}
		}
	}

	// os.LookupEnv returns false if var is not set
	env, ok := os.LookupEnv(key)
	switch {
	case !ok:
		slog.Info(fmt.Sprintf("env var <%v> not set up. using default <%v>", key, def))
		return def, nil
	case env == "":
		slog.Info(fmt.Sprintf("env var <%v> is set up, but empty. using default <%v>", key, def))
		return def, nil
	}

	switch any(def).(type) {
	case string:
		v = any(env).(T)
	case int:
		i, err := strconv.Atoi(env)
		if err != nil {
			return v, err
		}
		v = any(i).(T)
	case bool:
		b, err := strconv.ParseBool(env)
		if err != nil {
			return v, err
		}
		v = any(b).(T)
	}
	return v, nil
}

func setupLogger() *lumberjack.Logger {
	filename, err := getEnvOrDefaultString("FILENAME", "/var/log/openwrt/openwrt.log")
	if err != nil {
		slogFatal(err.Error())
	}

	maxSize, err := getEnvOrDefaultInt("MAXSIZE", 0)
	if err != nil {
		slogFatal(err.Error())
	}

	maxAge, err := getEnvOrDefaultInt("MAXAGE", 180)
	if err != nil {
		slogFatal(err.Error())
	}

	maxBackups, err := getEnvOrDefaultInt("MAXBACKUP", 0)
	if err != nil {
		slogFatal(err.Error())
	}

	compress, err := getEnvOrDefaultBool("COMPRESS", false)
	if err != nil {
		slogFatal(err.Error())
	}

	useLocalTime, err := getEnvOrDefaultBool("USELOCALTIME", true)
	if err != nil {
		slogFatal(err.Error())
	}

	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		LocalTime:  useLocalTime,
		Compress:   compress,
	}
}

func setupConfig() *Config {
	port, err := getEnvOrDefaultString("PORT", "8080")
	if err != nil {
		slogFatal(err.Error())
	}

	protocol, err := getEnvOrDefaultString("PROTOCOL", "tcp")
	if err != nil {
		slogFatal(err.Error())
	}

	address, err := getEnvOrDefaultString("ADDRESS", "localhost")
	if err != nil {
		slogFatal(err.Error())
	}

	return &Config{
		port:     port,
		logger:   setupLogger(),
		protocol: protocol,
		address:  address,
	}
}

// slog equivalent of log.Fatal()
func slogFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
