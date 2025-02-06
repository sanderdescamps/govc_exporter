package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualMachineSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.VirtualMachine]
	Refreshable
	Cleanable
}

func NewVirtualMachineSensor(scraper *VCenterScraper) *VirtualMachineSensor {
	return &VirtualMachineSensor{
		BaseSensor: BaseSensor[types.ManagedObjectReference, mo.VirtualMachine]{
			cache:   make(map[types.ManagedObjectReference]*CacheItem[mo.VirtualMachine]),
			scraper: scraper,
			metrics: nil,
		},
	}
}

func (s *VirtualMachineSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	var hostRefs []types.ManagedObjectReference
	hostRefs, err := func() ([]types.ManagedObjectReference, error) {
		resultChan := make(chan *[]types.ManagedObjectReference)
		quitChan := make(chan bool)
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					if s.scraper.Host != nil {
						refs := s.scraper.Host.GetAllRefs()
						if refs != nil {
							resultChan <- &refs
							break
						}
						logger.Info("VM sensor is waiting for hosts")
					} else {
						logger.Info("No host sensor found")
					}

				case <-quitChan:
					return
				}
			}
		}()

		select {
		case <-time.After(10 * time.Second):
			quitChan <- true
			logger.Warn("Scraper can not request virtual machines without hosts")
			return nil, fmt.Errorf("waiting for hosts timeout")
		case refs := <-resultChan:
			return *refs, nil
		}
	}()

	if err != nil {
		s.setMetrics(&SensorMetric{
			Name:   "vm",
			Status: false,
		})
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(hostRefs))
	resultChan := make(chan *[]mo.VirtualMachine, len(hostRefs))
	errChan := make(chan error, len(hostRefs))
	metricChan := make(chan SensorMetric, len(hostRefs))
	for _, host := range hostRefs {
		go func() {
			t1 := time.Now()
			client, release, err := s.scraper.clientPool.Acquire()
			if err != nil {
				errChan <- err
				return
			}
			defer release()
			defer wg.Done()
			t2 := time.Now()
			m := view.NewManager(client.Client)
			v, err := m.CreateContainerView(
				ctx,
				host,
				[]string{"VirtualMachine"},
				true,
			)
			if err != nil {
				errChan <- err
				return
			}
			defer v.Destroy(ctx)

			var items []mo.VirtualMachine
			err = v.Retrieve(
				context.Background(),
				[]string{"VirtualMachine"},
				[]string{
					"name",
					"config",
					//"datatore",
					"guest",
					"guestHeartbeatStatus",
					// "network",
					"parent",
					// "resourceConfig",
					"resourcePool",
					"runtime",
					// "snapshot",
					"summary",
				},
				&items,
			)
			t3 := time.Now()
			if err != nil {
				errChan <- err
			} else {
				resultChan <- &items
				metricChan <- SensorMetric{
					Name:           "vm",
					QueryTime:      t3.Sub(t2),
					ClientWaitTime: t2.Sub(t1),
					Status:         true,
				}
			}
		}()
	}

	readyChan := make(chan bool)
	failed := false
	var allClientWaitTimes []time.Duration
	var allQueryTimes []time.Duration
	go func() {
		for {
			select {
			case r := <-resultChan:
				for _, vm := range *r {
					s.Update(vm.Self, &vm)
				}
			case m := <-metricChan:
				allClientWaitTimes = append(allClientWaitTimes, m.ClientWaitTime)
				allQueryTimes = append(allQueryTimes, m.QueryTime)
			case err := <-errChan:
				logger.Error("failed to get vm's from host", "err", err)
				failed = true
			case <-readyChan:
				return
			}
		}
	}()

	wg.Wait()
	readyChan <- true

	s.setMetrics(&SensorMetric{
		Name:           "vm",
		QueryTime:      avgDuration(allClientWaitTimes),
		ClientWaitTime: avgDuration(allQueryTimes),
		Status:         !failed,
	})

	return nil
}

func (s *VirtualMachineSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	s.BaseSensor.Clean(maxAge, logger)
}
