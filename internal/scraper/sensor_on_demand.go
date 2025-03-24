package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type OnDemandSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ManagedEntity]
	Cleanable
	logger *slog.Logger
}

func NewOnDemandSensor(scraper *VCenterScraper, logger *slog.Logger) *OnDemandSensor {
	sensor := &OnDemandSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ManagedEntity](
			scraper,
		),
		logger: logger,
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration("sensor.on_demand.client_wait_time", 10)
	sensor.metrics.QueryTime = NewSensorMetricDuration("sensor.on_demand.query_time", 10)
	sensor.metrics.Status = NewSensorMetricStatus("sensor.on_demand.status", false)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return sensor
}

func (o *OnDemandSensor) Get(ref types.ManagedObjectReference) *mo.ManagedEntity {
	cacheEntity := func() *mo.ManagedEntity {
		o.lock.Lock()
		defer o.lock.Unlock()
		if entiry, ok := o.cache[ref]; ok {
			return entiry.Item
		} else {
			return nil
		}
	}()

	if cacheEntity != nil {
		o.logger.Debug("Found entity in cache", "entity", cacheEntity.Self.Encode())
		return cacheEntity
	}

	t1 := time.Now()
	client, release, err := o.scraper.clientPool.Acquire()
	if err != nil {
		o.logger.Error("Failed to acquire a client from pool", "err", err)
		return nil
	}
	defer release()

	t2 := time.Now()
	ctx := context.Background()
	pc := property.DefaultCollector(client.Client)
	o.logger.Debug("on_demand sensor query", "ref", ref)
	var entity mo.ManagedEntity
	err = pc.RetrieveOne(ctx, ref, []string{"name", "parent"}, &entity)

	t3 := time.Now()
	o.metrics.ClientWaitTime.Update(t2.Sub(t1))
	o.metrics.QueryTime.Update(t3.Sub(t2))
	o.metrics.Status.Success()
	if err != nil {
		o.metrics.Status.Fail()
		o.logger.Error("Failed to get on_demand object", "ref", ref, "err", err)
	}

	o.Update(ref, &entity)

	return &entity
}

func (s *OnDemandSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
}
