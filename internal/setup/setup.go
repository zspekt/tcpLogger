package setup

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/zspekt/tcpLogger/internal/utils"
)

type Cfg struct {
	Port     string
	Protocol string
	Address  string
	Logger   *lumberjack.Logger
}

type ArgError struct {
	Err   string
	Param []string
}

func (e *ArgError) Is(target error) bool {
	if e.Error() != target.Error() {
		return false
	}

	return true
}

func (e *ArgError) Error() string {
	return e.Err + ": " + fmt.Sprint(e.Param)
}

func getEnvOrDefaultLogLevel(key string, def slog.Level) (slog.Level, error) {
	v, err := getEnvOrDefaultGen(key, def)
	if err != nil {
		return 0, err
	}
	return v, nil
}

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

func getEnvOrDefaultGen[T string | bool | int | slog.Level](key string, def T) (T, error) {
	var v T

	// if this is a string..
	if defStr, ok := any(def).(string); ok {
		switch {

		// we can only check for zero values on strings
		// as both a 0 int, and a false bool aren't
		// invalid in this situation
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
		slog.Info(fmt.Sprintf("env var <%v> not set. using default <%v>", key, def))
		return def, nil
	case env == "":
		slog.Info(fmt.Sprintf("env var <%v> is set, but empty. using default <%v>", key, def))
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
	case slog.Level:
		levelMapper := map[string]slog.Level{
			"DEBUG": slog.LevelDebug,
			"INFO":  slog.LevelInfo,
			"WARN":  slog.LevelWarn,
			"ERROR": slog.LevelError,
		}

		v, ok := levelMapper[env]
		if !ok {
			slog.Error(fmt.Sprintf("invalid log level <%v> defaulting to <%v>", env, any(def).(slog.Level).Level().String()))
			return def, nil
		}
		return any(v).(T), nil
	}
	return v, nil
}

func logger() *lumberjack.Logger {
	const (
		defFilename     string = "/var/log/openwrt/openwrt.log"
		defMaxSize      int    = 0
		defMaxAge       int    = 180
		defMaxBackups   int    = 0
		defCompress     bool   = false
		defUseLocalTime bool   = true
	)

	filename, err := getEnvOrDefaultString("FILENAME", defFilename)
	utils.Must(err)

	maxSize, err := getEnvOrDefaultInt("MAXSIZE", defMaxSize)
	utils.Must(err)

	maxAge, err := getEnvOrDefaultInt("MAXAGE", defMaxAge)
	utils.Must(err)

	maxBackups, err := getEnvOrDefaultInt("MAXBACKUP", defMaxBackups)
	utils.Must(err)

	compress, err := getEnvOrDefaultBool("COMPRESS", defCompress)
	utils.Must(err)

	useLocalTime, err := getEnvOrDefaultBool("USELOCALTIME", defUseLocalTime)
	utils.Must(err)

	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		LocalTime:  useLocalTime,
		Compress:   compress,
	}
}

func Config() *Cfg {
	const (
		defLogLevel slog.Level = slog.LevelInfo
		defPort     string     = "8080"
		defProtocol string     = "tcp"
		defAddress  string     = "0.0.0.0"
	)

	// https://stackoverflow.com/a/76970969
	l := new(slog.LevelVar)
	l.Set(slog.LevelInfo) // this is how you dynamically the log level
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l})),
	)

	logLevel, err := getEnvOrDefaultLogLevel("LOGLEVEL", defLogLevel)
	utils.Must(err)

	if logLevel != slog.LevelInfo {
		l.Set(logLevel)
	}

	port, err := getEnvOrDefaultString("PORT", defPort)
	utils.Must(err)

	protocol, err := getEnvOrDefaultString("PROTOCOL", defProtocol)
	utils.Must(err)

	address, err := getEnvOrDefaultString("ADDRESS", defAddress)
	utils.Must(err)

	return &Cfg{
		Port:     port,
		Protocol: protocol,
		Address:  address,
		Logger:   logger(),
	}
}
