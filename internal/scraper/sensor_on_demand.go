package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/helper"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type OnDemandRequest struct {
	ReqRef   types.ManagedObjectReference
	RespChan chan *mo.ManagedEntity
}

func NewOnDemandRequest(ref types.ManagedObjectReference) *OnDemandRequest {
	return &OnDemandRequest{
		ReqRef:   ref,
		RespChan: make(chan *mo.ManagedEntity),
	}
}

type OnDemandSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.ManagedEntity]
	AutoCleanSensor
	started *helper.StartedCheck
	reqChan chan *OnDemandRequest
	stop    chan bool
}

func NewOnDemandSensor(scraper *VCenterScraper, config SensorConfig) *OnDemandSensor {
	var sensor OnDemandSensor
	sensor = OnDemandSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.ManagedEntity](
			"on_demand",
			"OnDemandSensor",
			helper.NewMatcher("on_demand"),
			scraper,
		),
		AutoCleanSensor: *NewAutoCleanSensor(&sensor, config),
		started:         helper.NewStartedCheck(),
		reqChan:         make(chan *OnDemandRequest),
		stop:            make(chan bool),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration(sensor.Kind(), "client_wait_time", 10)
	sensor.metrics.QueryTime = NewSensorMetricDuration(sensor.Kind(), "query_time", 10)
	sensor.metrics.Status = NewSensorMetricStatus(sensor.Kind(), "status", false)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return &sensor
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
		return cacheEntity
	}

	req := NewOnDemandRequest(ref)
	o.reqChan <- req
	entity := <-req.RespChan
	if entity != nil {
		o.Update(ref, entity)
	}
	return entity
}

func (o *OnDemandSensor) query(ctx context.Context, ref types.ManagedObjectReference) *mo.ManagedEntity {
	t1 := time.Now()
	client, release, err := o.scraper.clientPool.Acquire()
	if err != nil {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Error("Failed to acquire a client from pool", "err", err)
		}
		return nil
	}
	defer release()

	t2 := time.Now()
	pc := property.DefaultCollector(client.Client)
	if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
		logger.Debug("on_demand sensor query", "ref", ref)
	}
	var entity mo.ManagedEntity
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = pc.RetrieveOne(ctxWithTimeout, ref, []string{"name", "parent"}, &entity)

	t3 := time.Now()
	o.metrics.ClientWaitTime.Update(t2.Sub(t1))
	o.metrics.QueryTime.Update(t3.Sub(t2))
	o.metrics.Status.Success()
	if err != nil {
		o.metrics.Status.Fail()
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Error("Failed to get on_demand object", "ref", ref, "err", err)
		}
	}

	return &entity
}

func (o *OnDemandSensor) Start(ctx context.Context) error {
	if !o.started.IsStarted() {
		go func() {
			o.started.Started()
			for o.started.IsStarted() {
				select {
				case req := <-o.reqChan:
					entiry := o.query(ctx, req.ReqRef)
					req.RespChan <- entiry
				case <-ctx.Done():
					o.started.Stopped()
				case <-o.stop:
					o.started.Stopped()
				}
			}
		}()
	} else {
		if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
			logger.Warn("Sensor already started", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
		}
	}
	return nil
}

func (o *OnDemandSensor) Stop(ctx context.Context) {
	if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
		logger.Info("stopping sensor...", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
	}
	o.stop <- true
	o.started.Stopped()
	if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
		logger.Info("sensor stopped", "sensor_name", o.sensor.Name(), "sensor_kind", o.sensor.Kind())
	}

}

func (o *OnDemandSensor) WaitTillStartup() {
	o.started.Wait()
}

func (o *OnDemandSensor) TriggerInstantRefresh(ctx context.Context) error {
	return nil
}

func (o *OnDemandSensor) Refresh(ctx context.Context) error {
	return nil
}
