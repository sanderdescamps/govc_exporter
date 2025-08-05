package scraper

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/sanderdescamps/govc_exporter/internal/database"
	memory_db "github.com/sanderdescamps/govc_exporter/internal/database/memory"
	"github.com/sanderdescamps/govc_exporter/internal/pool"
	metricshelper "github.com/sanderdescamps/govc_exporter/internal/scraper/metrics_helper"
)

// type Error struct {
// 	err  error
// 	args []any
// }

// func (e *Error) AddArg(key string, value any) *Error {
// 	e.args = append(e.args, key, value)
// 	return e
// }

type VCenterScraper struct {
	clientPool pool.VCenterPool
	config     Config
	DB         database.Database
	MetricsDB  database.MetricDB

	Host             Sensor
	HostPerf         Sensor
	Cluster          Sensor
	ComputeResources Sensor
	VM               Sensor
	VMPerf           Sensor
	Datastore        Sensor
	SPOD             Sensor
	ResourcePool     Sensor
	Tags             Sensor
	Datacenter       Sensor
	Folder           Sensor
	// Remain           *OnDemandSensor
}

func NewVCenterScraper(ctx context.Context, conf Config, logger *slog.Logger) (*VCenterScraper, error) {
	pool := pool.NewVCenterThrottlePool(
		conf.Endpoint(),
		conf.Username,
		conf.Password,
		conf.ClientPoolSize,
	)

	err := pool.Init()
	if err != nil {
		return nil, err
	}
	// pool.StartAuthRefresher(ctx, 5*time.Minute)

	var db database.Database = memory_db.NewDB()
	db.Connect(ctx)

	var metricsDb database.MetricDB = memory_db.NewMetricsDB()
	metricsDb.Connect(ctx)

	scraper := VCenterScraper{
		clientPool: pool,
		config:     conf,
		DB:         db,
		MetricsDB:  metricsDb,
	}

	if conf.Cluster.Enabled {
		scraper.Cluster = NewClusterSensor(&scraper, conf.Cluster, logger)
	} else {
		scraper.Cluster = NewNullSensor(CLUSTER_SENSOR_NAME)
	}

	if conf.ComputeResource.Enabled {
		scraper.ComputeResources = NewComputeResourceSensor(&scraper, conf.ComputeResource, logger)
	} else {
		scraper.ComputeResources = NewNullSensor(COMPUTE_RESOURCE_SENSOR_NAME)
	}

	if conf.Datastore.Enabled {
		scraper.Datastore = NewDatastoreSensor(&scraper, conf.Datastore, logger)
	} else {
		scraper.Datastore = NewNullSensor(DATASTORE_SENSOR_NAME)
	}

	if conf.Host.Enabled {
		scraper.Host = NewHostSensor(&scraper, conf.Host, logger)
		if conf.HostPerf.Enabled {
			scraper.HostPerf = NewHostPerfSensor(&scraper, conf.HostPerf, logger)
		} else {
			scraper.HostPerf = NewNullSensor(HOST_PERF_SENSOR_NAME)
		}
	} else {
		scraper.Host = NewNullSensor(HOST_SENSOR_NAME)
		scraper.HostPerf = NewNullSensor(HOST_PERF_SENSOR_NAME)
	}

	if conf.ResourcePool.Enabled {
		scraper.ResourcePool = NewResourcePoolSensor(&scraper, conf.ResourcePool, logger)
	} else {
		scraper.ResourcePool = NewNullSensor(RESOURCE_POOL_SENSOR_NAME)
	}

	if conf.Spod.Enabled {
		scraper.SPOD = NewStoragePodSensor(&scraper, conf.Spod, logger)
	} else {
		scraper.SPOD = NewNullSensor(STORAGE_POD_SENSOR_NAME)
	}

	if conf.VirtualMachine.Enabled {
		scraper.VM = NewVirtualMachineSensor(&scraper, conf.VirtualMachine, logger)
		if conf.VirtualMachinePerf.Enabled {
			scraper.VMPerf = NewVMPerfSensor(&scraper, conf.VirtualMachinePerf, logger)
		} else {
			scraper.VMPerf = NewNullSensor(VM_PERF_SENSOR_NAME)
		}
	} else {
		scraper.VM = NewNullSensor(VM_SENSOR_NAME)
		scraper.VMPerf = NewNullSensor(VM_PERF_SENSOR_NAME)
	}

	if conf.Tags.Enabled {
		logger.Info("Create TagsSensor", "TagsCategoryToCollect", conf.Tags.CategoryToCollect)
		scraper.Tags = NewTagsSensor(&scraper, conf.Tags, logger)
	} else {
		scraper.Tags = NewNullSensor("TagsSensor")
	}

	if conf.Folder.Enabled {
		scraper.Folder = NewFolderSensor(&scraper, conf.Folder, logger)
	} else {
		scraper.Folder = NewNullSensor(FOLDER_SENSOR_NAME)
	}

	if conf.Datacenter.Enabled {
		scraper.Datacenter = NewDatacenterSensor(&scraper, conf.Datacenter, logger)
	} else {
		scraper.Datacenter = NewNullSensor(DATACENTER_SENSOR_NAME)
	}

	return &scraper, nil

}

var (
	ErrVCenterURLInvalid  = errors.New("could not parse url")
	ErrVCenterConnectFail = errors.New("cannot connect to vcenter")
)

func (c *VCenterScraper) tcpConnectStatus() error {
	baseURL, err := c.config.URL()
	if err != nil {
		return ErrVCenterURLInvalid
	}

	_, err = tcpConnectionCheck(baseURL.Host)
	if err != nil {
		return ErrVCenterConnectFail
	}

	return nil
}

func (c *VCenterScraper) CollectMetrics() []metricshelper.SensorMetric {
	metrics := []metricshelper.SensorMetric{}
	metrics = append(metrics, c.ScraperMetrics()...)
	for _, sensor := range c.SensorList() {
		metrics = append(metrics, sensor.GetLatestMetrics()...)
	}
	return metrics
}

func (c *VCenterScraper) ScraperMetrics() []metricshelper.SensorMetric {
	result := []metricshelper.SensorMetric{}

	err := c.tcpConnectStatus()
	if err == nil {
		result = append(result, metricshelper.SensorMetric{
			Sensor:     "scraper",
			MetricName: "tcp_connect_status",
			Unit:       "boolean",
			Value:      1,
		})
	} else if errors.Is(err, ErrVCenterURLInvalid) {
		result = append(result, metricshelper.SensorMetric{
			Sensor:     "scraper",
			MetricName: "tcp_connect_status",
			Unit:       "boolean",
			Value:      0,
		})
	} else if errors.Is(err, ErrVCenterConnectFail) {
		result = append(result, metricshelper.SensorMetric{
			Sensor:     "scraper",
			MetricName: "tcp_connect_status",
			Unit:       "boolean",
			Value:      0,
		})
	} else {
		result = append(result, metricshelper.SensorMetric{
			Sensor:     "scraper",
			MetricName: "tcp_connect_status",
			Unit:       "boolean",
			Value:      0,
		})
	}
	return result
}

func (c *VCenterScraper) SensorList() []Sensor {
	return []Sensor{
		c.Folder,
		c.Datacenter,
		c.Cluster,
		c.ComputeResources,
		c.SPOD,
		c.Datastore,
		c.ResourcePool,
		c.Tags,
		c.Host,
		c.VM,
		c.HostPerf,
		c.VMPerf,
	}
}

func (c *VCenterScraper) Start(ctx context.Context, logger *slog.Logger) error {
	err := c.tcpConnectStatus()
	if err != nil {
		return fmt.Errorf("cannot connect to vcenter: %w", err)
	}

	// Start all sensors
	sensors := c.SensorList()
	for i, sensor := range sensors {
		if err := ctx.Err(); err != nil {
			break
		} else if sensor.Enabled() {
			logger.Info(fmt.Sprintf("Initialize sensor... [%d/%d]", i+1, len(sensors)), "sensor_kind", sensor.Kind())
			err := sensor.Init(ctx, c)
			if err != nil {
				logger.Error("Failed init sensor", "sensor_kind", sensor.Kind(), "err", err)
			} else {
				logger.Debug("Init Sensor successful", "sensor_kind", sensor.Kind())
			}
		} else {
			logger.Info(fmt.Sprintf("skip sensor [%d/%d]", i+1, len(sensors)), "sensor_kind", sensor.Kind())
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	logger.Info("All sensors initialized")

	wg := sync.WaitGroup{}
	for _, sensor := range sensors {
		if sensor.Enabled() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sensor.StartRefresher(ctx, c)
			}()
		}
	}

	wg.Wait()
	logger.Info("All sensors refreshers started")

	return nil
}

func (c *VCenterScraper) Stop(ctx context.Context, logger *slog.Logger) {
	for _, sensor := range c.SensorList() {
		if sensor.Enabled() {
			logger.Info("Stop sensor...", "sensor_kind", sensor.Kind())
			sensor.StopRefresher(ctx)
		} else {
			logger.Info("Skip sensor, sensor disabled", "sensor_kind", sensor.Kind())
		}
	}

	logger.Info("finish triggering termination of all sensors")

	c.clientPool.Destroy(ctx)
	if logger, ok := ctx.Value(ContextKeyScraperLogger{}).(*slog.Logger); ok {
		logger.Info("Close client pool")
	}
}

func (c *VCenterScraper) TriggerSensorRefreshByName(ctx context.Context, sensorName string) error {
	for _, sensor := range c.SensorList() {
		if sensor.Match(sensorName) {
			sensor.TriggerManualRefresh(ctx)
			return nil
		}
	}
	return ErrSensorNotFound
}

type ParentChain struct {
	DC           string
	Cluster      string
	ResourcePool string
	SPOD         string
	Chain        []string
}
