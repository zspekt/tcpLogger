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
	Logger   *lumberjack.Logger
	Protocol string
	Address  string
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

func logger() *lumberjack.Logger {
	filename, err := getEnvOrDefaultString("FILENAME", "/var/log/openwrt/openwrt.log")
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	maxSize, err := getEnvOrDefaultInt("MAXSIZE", 0)
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	maxAge, err := getEnvOrDefaultInt("MAXAGE", 180)
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	maxBackups, err := getEnvOrDefaultInt("MAXBACKUP", 0)
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	compress, err := getEnvOrDefaultBool("COMPRESS", false)
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	useLocalTime, err := getEnvOrDefaultBool("USELOCALTIME", true)
	if err != nil {
		utils.SlogFatal(err.Error())
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

func Config() *Cfg {
	port, err := getEnvOrDefaultString("PORT", "8080")
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	protocol, err := getEnvOrDefaultString("PROTOCOL", "tcp")
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	address, err := getEnvOrDefaultString("ADDRESS", "localhost")
	if err != nil {
		utils.SlogFatal(err.Error())
	}

	return &Cfg{
		Port:     port,
		Logger:   logger(),
		Protocol: protocol,
		Address:  address,
	}
}
