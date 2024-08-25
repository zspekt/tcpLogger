package main

import (
	"fmt"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	port     string
	logger   *lumberjack.Logger
	protocol string
	address  string
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
