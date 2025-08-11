package scraper

import (
	"errors"
	"fmt"
	"strings"
)

var ErrSensorAlreadyRunning = errors.New("sensor already running")
var ErrSensorAlreadyStarted = errors.New("sensor already started")
var ErrSensorInitTimeout = errors.New("sensor init timeout")
var ErrSensorCientFailed = errors.New("sensor failed to get client")
var ErrSensorNotFound = errors.New("sensor not found")

type SensorError struct {
	msg  string
	args []any
}

func NewSensorError(msg string, args ...any) *SensorError {
	return &SensorError{
		msg:  msg,
		args: args,
	}
}

func (e *SensorError) Error() string {
	extra := []string{}

	for i := 0; i < len(e.args); i += 2 {
		if i+1 < len(e.args) {
			extra = append(extra, fmt.Sprintf("%v=%v", e.args[i], e.args[i+1]))
		} else {
			extra = append(extra, fmt.Sprintf("%v", e.args[i]))
		}
	}
	return e.msg + strings.Join(extra, " ")
}
