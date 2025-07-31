package scraper

import "errors"

var ErrSensorAlreadyRunning = errors.New("Sensor already running")
var ErrSensorCientFailed = errors.New("Sensor failed to get client")
var ErrSensorNotFound = errors.New("Sensor not found")

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
	return e.msg
}
