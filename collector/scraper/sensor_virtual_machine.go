package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/sanderdescamps/govc_exporter/collector/helper"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualMachineSensor struct {
	BaseSensor[types.ManagedObjectReference, mo.VirtualMachine]
	Refreshable
	helper.Matchable
}

func NewVirtualMachineSensor(scraper *VCenterScraper) *VirtualMachineSensor {
	sensor := &VirtualMachineSensor{
		BaseSensor: *NewBaseSensor[types.ManagedObjectReference, mo.VirtualMachine](
			scraper,
		),
	}
	sensor.metrics.ClientWaitTime = NewSensorMetricDuration("sensor.vm.client_wait_time", 10)
	sensor.metrics.QueryTime = NewSensorMetricDuration("sensor.vm.query_time", 10)
	sensor.metrics.Status = NewSensorMetricStatus("sensor.vm.status", true)
	scraper.RegisterSensorMetric(
		&sensor.metrics.ClientWaitTime.SensorMetric,
		&sensor.metrics.QueryTime.SensorMetric,
		&sensor.metrics.Status.SensorMetric,
	)
	return sensor
}

func (s *VirtualMachineSensor) Refresh(ctx context.Context, logger *slog.Logger) error {
	sensorKind := reflect.TypeOf(s).String()
	if hasLock := s.refreshLock.TryLock(); hasLock {
		defer s.refreshLock.Unlock()
		return s.unsafeRefresh(ctx, logger)
	} else {
		logger.Info("Sensor Refresh already running", "sensor_type", sensorKind)
	}
	return nil
}

func (s *VirtualMachineSensor) unsafeRefresh(ctx context.Context, logger *slog.Logger) error {
	s.metrics.Status.Reset()
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
		s.metrics.Status.Fail()
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(hostRefs))
	resultChan := make(chan *[]mo.VirtualMachine, len(hostRefs))
	errChan := make(chan error, len(hostRefs))
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
				s.metrics.ClientWaitTime.Update(t2.Sub(t1))
				s.metrics.QueryTime.Update(t3.Sub(t2))
				s.metrics.Status.Success()
			}
		}()
	}

	readyChan := make(chan bool)
	allVMs := map[types.ManagedObjectReference]*CacheItem[mo.VirtualMachine]{}
	go func() {
		for {
			select {
			case r := <-resultChan:
				for _, vm := range *r {
					if entityA, exist := allVMs[vm.Self]; exist {
						timestampA := func() time.Time {
							if entityA.Item.Config != nil {
								timestamp, err := time.Parse(time.RFC3339Nano, entityA.Item.Config.ChangeVersion)
								if err == nil {
									return timestamp
								}
							}
							return time.Now()
						}()

						timestampB := func() time.Time {
							if vm.Config != nil {
								timestamp, err := time.Parse(time.RFC3339Nano, vm.Config.ChangeVersion)
								if err == nil {
									return timestamp
								}
							}
							return time.Now()
						}()
						// time comparison to prevent inconsistent data during vmotion
						if timestampB.After(timestampA) {
							allVMs[vm.Self] = NewCacheItem(&vm)
						}
					} else {
						allVMs[vm.Self] = NewCacheItem(&vm)
					}
				}
			case err := <-errChan:
				logger.Error("failed to get vm's from host", "err", err)
				s.metrics.Status.Fail()
			case <-readyChan:
				return
			}
		}
	}()

	wg.Wait()
	readyChan <- true

	for ref, cacheItem := range allVMs {
		s.UpdateCacheItem(ref, cacheItem)
	}

	return nil
}

func (s *VirtualMachineSensor) GetHostVMs(ref types.ManagedObjectReference) []types.ManagedObjectReference {
	vmsOnHost := []types.ManagedObjectReference{}
	for _, vm := range s.GetAll() {
		if vm.Runtime.Host != nil && *vm.Runtime.Host == ref {
			vmsOnHost = append(vmsOnHost, vm.Self)
		}
	}
	return vmsOnHost
}

func (s *VirtualMachineSensor) Name() string {
	return "vm"
}

func (s *VirtualMachineSensor) Match(name string) bool {
	return helper.NewMatcher("vm", "virtual_machine", "virtualmachine").Match(name)
}
