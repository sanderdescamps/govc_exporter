package scraper

import (
	"context"
	"time"

	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
	"github.com/vmware/govmomi/view"
)

type BaseSensor struct {
	moType       string
	moProperties []string
}

func NewBaseSensor(moType string, moProperties []string) *BaseSensor {
	return &BaseSensor{
		moType:       moType,
		moProperties: moProperties,
	}
}

func (s *BaseSensor) baseRefresh(ctx context.Context, scraper *VCenterScraper, res interface{}) (metricshelper.RefreshStats, error) {
	t1 := time.Now()
	client, release, err := scraper.clientPool.AcquireWithContext(ctx)
	if err != nil {
		return metricshelper.RefreshStats{
			Failed: true,
		}, err
	}
	defer release()
	t2 := time.Now()

	m := view.NewManager(client.Client)
	v, err := m.CreateContainerView(
		ctx,
		client.ServiceContent.RootFolder,
		[]string{s.moType},
		true,
	)
	if err != nil {
		return metricshelper.RefreshStats{
			Failed: true,
		}, err
	}
	defer v.Destroy(ctx)

	err = v.Retrieve(
		ctx,
		[]string{s.moType},
		s.moProperties,
		res,
	)
	t3 := time.Now()
	return metricshelper.RefreshStats{
		ClientWaitTime: t2.Sub(t1),
		QueryTime:      t3.Sub(t2),
		Failed: func() bool {
			return err != nil
		}(),
	}, err
}
