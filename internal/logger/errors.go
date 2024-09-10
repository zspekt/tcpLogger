package logger

import "errors"

var (
	NetTimeoutError  error = errors.New("timeout waiting for connection")
	ReadTimeoutError error = errors.New("timeout waiting for reader")
)
