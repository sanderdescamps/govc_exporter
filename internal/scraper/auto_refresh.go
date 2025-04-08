package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
)

const (
	// factor to multiply the refresh interval to get the timeout of the sensor refresh
	RefreshTimeoutFactor = 3
)

type AutoRefreshSensor struct {
	refreshInterval time.Duration
	sensor          Sensor
	config          SensorConfig
	started         *helper.StartedCheck
	triggerRefresh  chan bool
}

func NewAutoRefreshSensor(sensor Sensor, config SensorConfig) *AutoRefreshSensor {
	refreshInterval := config.RefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 60 * time.Second
	}

	return &AutoRefreshSensor{
		refreshInterval: refreshInterval,
		sensor:          sensor,
		config:          config,
		started:         helper.NewStartedCheck(),
		triggerRefresh:  make(chan bool),
	}
}

func (o *AutoRefreshSensor) Start(ctx context.Context) error {
	if !o.started.IsStarted() {
		t1 := time.Now()
		err := o.sensor.Refresh(ctx)
		if err != nil {
			return fmt.Errorf("failed to run initial sensor refresh : %w", err)
		}
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			msg := fmt.Sprintf("initial refresh successful (%dms)", time.Since(t1).Milliseconds())
			logger.Debug(msg, "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
		}

		go func() {
			refreshFunc := func(refreshCtx context.Context) {
				err := (o.sensor).Refresh(refreshCtx)
				if logger, ok := refreshCtx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
					if err == nil {
						logger.Info("Refresh successfull", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
					} else {
						logger.Warn("Failed to refresh sensor", "err", err.Error(), "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
					}
				}
			}
			ticker := time.NewTicker(o.refreshInterval)
			for {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, o.refreshInterval*RefreshTimeoutFactor)
				select {
				case <-ticker.C:
					refreshFunc(ctxWithTimeout)
				case <-o.triggerRefresh:
					refreshFunc(ctxWithTimeout)
				case <-ctxWithTimeout.Done():
					if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
						msg := fmt.Sprintf("sensor refresh timeout after %dsec", int(o.refreshInterval.Seconds())*RefreshTimeoutFactor)
						logger.Warn(msg, "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
					}
				case <-ctx.Done():
					ticker.Stop()
					cancel()
					close(o.triggerRefresh)
					if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
						logger.Info("sensor refresh stopped", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
					}
					return
				}
				cancel()
			}
		}()

	} else {
		return fmt.Errorf("Sensor already started")
	}
	o.started.Started()
	return nil
}

func (o *AutoRefreshSensor) TriggerRefresh(ctx context.Context) error {
	done := make(chan bool)
	go func() {
		o.triggerRefresh <- true
		done <- true
	}()

	select {
	case <-done:
		return nil
	case <-time.After(60 * time.Second):
		return fmt.Errorf("manual refresh timeout")
	case <-ctx.Done():
		return fmt.Errorf("manual refresh timeout")
	}
}

func (o *AutoRefreshSensor) WaitTillStartup() {
	o.started.Wait()
}
