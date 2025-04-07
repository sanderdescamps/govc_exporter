package scraper

import (
	"context"
	"log/slog"
	"time"
)

type AutoCleanSensor struct {
	sensor CleanOnlySensor
	config SensorConfig
}

func NewAutoCleanSensor(sensor CleanOnlySensor, config SensorConfig) *AutoCleanSensor {
	return &AutoCleanSensor{
		sensor: sensor,
		config: config,
	}
}

func (o *AutoCleanSensor) Start(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(o.config.CleanInterval)
		for {
			select {
			case <-ticker.C:
				o.sensor.Clean(ctx, o.config.MaxAge)
				if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
					logger.Debug("clean successfull", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
				}
			case <-ctx.Done():
				ticker.Stop()
				if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
					logger.Info("cleanup ticker stopped", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
				}
				return
			}
		}
	}()
	return nil
}
