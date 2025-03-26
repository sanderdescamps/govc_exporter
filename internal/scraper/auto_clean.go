package scraper

import (
	"context"
	"log/slog"
	"time"
)

type AutoCleanSensor struct {
	sensor CleanOnlySensor
	ticker *time.Ticker
	config SensorConfig
}

func NewAutoCleanSensor(sensor CleanOnlySensor, config SensorConfig) *AutoCleanSensor {
	return &AutoCleanSensor{
		sensor: sensor,
		ticker: time.NewTicker(config.CleanInterval),
		config: config,
	}
}

func (o *AutoCleanSensor) Start(ctx context.Context, logger *slog.Logger) {
	go func() {
		for ; true; <-o.ticker.C {
			o.sensor.Clean(o.config.MaxAge, logger)
			logger.Debug("clean successfull", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
		}
	}()

}

func (o *AutoCleanSensor) Stop(ctx context.Context, logger *slog.Logger) {
	logger.Info("stopping cleanup ticker...", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
	o.ticker.Stop()
	logger.Info("cleanup ticker stopped", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
}
