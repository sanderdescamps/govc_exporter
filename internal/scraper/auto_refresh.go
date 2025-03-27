package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

type AutoRefreshSensor struct {
	rTicker *time.Ticker
	cTicker *time.Ticker
	sensor  Sensor
	config  SensorConfig
	started *helper.StartedCheck
}

func NewAutoRefreshSensor(sensor Sensor, config SensorConfig) *AutoRefreshSensor {
	return &AutoRefreshSensor{
		rTicker: time.NewTicker(config.RefreshInterval),
		cTicker: time.NewTicker(config.CleanInterval),
		sensor:  sensor,
		config:  config,
		started: helper.NewStartedCheck(),
	}
}

func (o *AutoRefreshSensor) Start(ctx context.Context, logger *slog.Logger) {
	if !o.started.IsStarted() {
		go func() {
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
			for ; true; <-o.rTicker.C {
				err := (o.sensor).Refresh(ctx, logger)
				if err == nil {
					o.started.Started()
					logger.Info("Refresh successfull", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
				} else {
					logger.Warn("Failed to refresh sensor", "err", err.Error(), "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
				}
			}
		}()
	}
}

func (o *AutoRefreshSensor) Stop(ctx context.Context, logger *slog.Logger) {
	logger.Info("stopping refresh rTicker...", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
	o.rTicker.Stop()
	o.started.Stopped()
	logger.Info("refresh rTicker stopped", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
}

func (o *AutoRefreshSensor) WaitTillStartup() {
	o.started.Wait()
}
