package scraper

import (
	"log/slog"
	"reflect"
	"time"
)

type Cleanable interface {
	Clean(maxAge time.Duration)
}

type AutoClean struct {
	sensor        Cleanable
	maxAge        time.Duration
	cleanupTicker *time.Ticker
}

func NewAutoClean(sensor Cleanable, interval time.Duration, maxAge time.Duration) *AutoClean {
	return &AutoClean{
		sensor:        sensor,
		maxAge:        maxAge,
		cleanupTicker: time.NewTicker(interval),
	}
}

func (o *AutoClean) Start(logger *slog.Logger) {
	sensorKind := reflect.TypeOf(o.sensor).String()

	go func() {
		for ; true; <-o.cleanupTicker.C {

			o.sensor.Clean(o.maxAge)
			logger.Debug("clean successfull", "sensor_type", sensorKind)

		}
	}()

}

func (o *AutoClean) Stop(logger *slog.Logger) {
	sensorKind := reflect.TypeOf(o.sensor).String()
	logger.Info("stopping cleanup ticker...", "sensor_type", sensorKind)
	o.cleanupTicker.Stop()
	logger.Info("cleanup ticker stopped", "sensor_type", sensorKind)
}
