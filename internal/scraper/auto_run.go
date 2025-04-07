package scraper

import (
	"context"
)

type Runnable interface {
	Start(context.Context) error
	WaitTillStartup()
}

type AutoRunSensor struct {
	AutoCleanSensor
	AutoRefreshSensor
}

func (o *AutoRunSensor) Start(ctx context.Context) error {
	refreshErr := o.AutoRefreshSensor.Start(ctx)
	if refreshErr != nil {
		return refreshErr
	}

	cleanErr := o.AutoCleanSensor.Start(ctx)
	if cleanErr != nil {
		return cleanErr
	}
	return nil
}

func NewAutoRunSensor(sensor Sensor, config SensorConfig) *AutoRunSensor {
	return &AutoRunSensor{
		AutoRefreshSensor: *NewAutoRefreshSensor(sensor, config),
		AutoCleanSensor:   *NewAutoCleanSensor(sensor, config),
	}
}
