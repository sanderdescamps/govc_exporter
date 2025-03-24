package scraper

import (
	"context"
	"log/slog"
	"math/rand"
	"reflect"
	"time"
)

type Refreshable interface {
	Refresh(context.Context, *slog.Logger) error
}

type RefreshConfig struct {
	RefreshInterval    int64
	CleanCacheInterval int64
	MaxAge             int64
}

type AutoRefresh struct {
	sensor        Refreshable
	refreshTicker *time.Ticker
}

func NewAutoRefresh(sensor Refreshable, interval time.Duration) *AutoRefresh {
	return &AutoRefresh{
		sensor:        sensor,
		refreshTicker: time.NewTicker(interval),
	}
}

func (o *AutoRefresh) Start(ctx context.Context, logger *slog.Logger) {
	sensorKind := reflect.TypeOf(o.sensor).String()

	go func() {
		//random sleep to prevent sensors refreshing at the same time
		time.Sleep(time.Duration(rand.Intn(3)) * time.Second)

		for ; true; <-o.refreshTicker.C {
			err := o.sensor.Refresh(ctx, logger)
			if err == nil {
				logger.Info("refresh successfull", "sensor_type", sensorKind)
			} else {
				logger.Warn("Failed to refresh sensor", "err", err.Error(), "sensor_type", sensorKind)
			}
		}
	}()
	logger.Info("Sensor start successfull", "sensor_type", sensorKind)
}

func (o *AutoRefresh) Stop(logger *slog.Logger) {
	sensorKind := reflect.TypeOf(o.sensor).String()
	logger.Info("stopping refresh ticker...", "sensor_type", sensorKind)
	o.refreshTicker.Stop()
	logger.Info("refresh ticker stopped", "sensor_type", sensorKind)
}
