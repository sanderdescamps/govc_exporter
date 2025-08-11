package scraper

import (
	"context"

	sensormetrics "github.com/sanderdescamps/govc_exporter/internal/scraper/sensor_metrics"
	"github.com/vmware/govmomi/view"
)

type BaseSensor struct {
	moType       string
	moProperties []string

	metricsCollector *sensormetrics.SensorMetricsCollector
	statusMonitor    *sensormetrics.StatusMonitor
}

func NewBaseSensor(moType string, moProperties []string, mc *sensormetrics.SensorMetricsCollector, sm *sensormetrics.StatusMonitor) *BaseSensor {
	return &BaseSensor{
		moType:           moType,
		moProperties:     moProperties,
		metricsCollector: mc,
		statusMonitor:    sm,
	}
}

func (s *BaseSensor) baseRefresh(ctx context.Context, scraper *VCenterScraper, res interface{}) error {
	sensorStopwatch := sensormetrics.NewSensorStopwatch()
	sensorStopwatch.Start()
	client, release, err := scraper.clientPool.AcquireWithContext(ctx)
	if err != nil {
		return err
	}
	defer release()
	sensorStopwatch.Mark1()

	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{s.moType},
		true,
	)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{s.moType},
		s.moProperties,
		res,
	)
	sensorStopwatch.Finish()
	s.metricsCollector.UploadStats(sensorStopwatch.GetStats())
	return err
}
