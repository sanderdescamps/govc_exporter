package scraper

import (
	"context"
	"log/slog"
)

type Runnable interface {
	Start(context.Context, *slog.Logger)
	Stop(context.Context, *slog.Logger)
}

type AutoRunSensor struct {
	AutoCleanSensor
	AutoRefreshSensor
}

func (o *AutoRunSensor) Start(ctx context.Context, logger *slog.Logger) {
	o.AutoRefreshSensor.Start(ctx, logger)
	o.AutoCleanSensor.Start(ctx, logger)
}

func (o *AutoRunSensor) Stop(ctx context.Context, logger *slog.Logger) {
	o.AutoRefreshSensor.Stop(ctx, logger)
	o.AutoCleanSensor.Stop(ctx, logger)
}

func NewAutoRunSensor(sensor Sensor, config SensorConfig) *AutoRunSensor {
	return &AutoRunSensor{
		AutoRefreshSensor: *NewAutoRefreshSensor(sensor, config),
		AutoCleanSensor:   *NewAutoCleanSensor(sensor, config),
	}
}
