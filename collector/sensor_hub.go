package collector

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
)

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
}

type SensorHub struct {
	ctx       context.Context
	cache     []any
	client    *govmomi.Client
	config    SensorHubConf
	jobTicker *time.Ticker
}

type SensorHubConf struct {
	SensorRefreshPeriod int    `json:"sensor_refresh_interval"`
	SensorHubJobPeriod  int    `json:"sensor_hub_job_period"`
	Endpoint            string `json:"endpoint"`
	Username            string `json:"username"`
	Password            string `json:"password"`
}

func (c SensorHubConf) URL() string {
	return c.Endpoint
}

type ContextKey string

const (
	SensorLoggerCtxKey ContextKey = "sensor-logger"
)

// func NexVCenterHub(ctx context.Context, conf SensorHubConf) (*SensorHub, error) {
// 	// logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
// 	ctx = context.WithValue(ctx, SensorLoggerCtxKey, logger)
// 	sh := &SensorHub{
// 		ctx:    ctx,
// 		config: conf,
// 	}

// 	sh.sensors = []*Sensor{}
// 	sh.sensors = append(sh.sensors, NewSensor(
// 		"govc_esx",
// 		NewHostRefreshFunc(sh.ctx, sh),
// 		time.Duration(time.Duration(sh.config.SensorRefreshPeriod)*time.Second),
// 	))
// 	sh.sensors = append(sh.sensors, NewSensor(
// 		"govc_ds",
// 		NewDatastoreRefreshFunc(sh.ctx, sh),
// 		time.Duration(time.Duration(sh.config.SensorRefreshPeriod)*time.Second),
// 	))
// 	sh.sensors = append(sh.sensors, NewSensor(
// 		"govc_spod",
// 		NewSpodRefreshFunc(sh.ctx, sh),
// 		time.Duration(time.Duration(sh.config.SensorRefreshPeriod)*time.Second),
// 	))

// 	return sh, nil
// }

func NewSensorHub(ctx context.Context, conf SensorHubConf) (*SensorHub, error) {
	// logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	ctx = context.WithValue(ctx, SensorLoggerCtxKey, logger)
	sh := SensorHub{
		ctx:    ctx,
		config: conf,
	}

	return &sh, nil
}

func (sh *SensorHub) Login() {
	// logger := sh.ctx.Value(SensorLoggerCtxKey).(*slog.Logger)
	urlString := sh.config.URL()
	u, err := soap.ParseURL(urlString)
	if err != nil {
		logger.Error("unable to parse", "url", urlString, "err", err)
		panic(err)
	}
	u.User = url.UserPassword(sh.config.Username, sh.config.Password)
	logger.Debug("connecting to", "url", urlString)
	client, err := govmomi.NewClient(sh.ctx, u, true)
	if err != nil {
		logger.Error("failed to login", "url", urlString, "err", err)
		panic(err)
	}
	sh.client = client
}

func (sh *SensorHub) Logout() {
	logger := sh.ctx.Value(SensorLoggerCtxKey).(*slog.Logger)
	err := sh.client.Logout(sh.ctx)
	if err != nil {
		logger.LogAttrs(sh.ctx, slog.LevelError, "logout error", slog.String("err", err.Error()))
	}
	sh.ctx.Done()
}

func (sh *SensorHub) GetClient() *govmomi.Client {
	return sh.client
}

// func (sh *SensorHub) NewHostSensor() (*Sensor, error) {
// 	hostRefreshFunc := NewHostRefreshFunc(sh.ctx, sh)

// 	s := Sensor{
// 		parentHub:       sh,
// 		refreshFunction: hostRefreshFunc,
// 		LastRefresh:     time.Date(0, 0, 0, 0, 0, 0, 0, time.Local),
// 		RefreshPeriod:   time.Duration(time.Duration(sh.config.SensorRefreshPeriod) * time.Second),
// 	}
// 	sh.sensors = append(sh.sensors, s)
// 	return &s, nil
// }

// func (sh *SensorHub) NewDatastoreSensor() (*Sensor, error) {
// 	hostRefreshFunc := NewDatastoreRefreshFunc(sh.ctx, sh)

// 	s := Sensor{
// 		parentHub:       sh,
// 		refreshFunction: hostRefreshFunc,
// 		LastRefresh:     time.Date(0, 0, 0, 0, 0, 0, 0, time.Local),
// 		RefreshPeriod:   time.Duration(time.Duration(sh.config.SensorRefreshPeriod) * time.Second),
// 	}
// 	sh.sensors = append(sh.sensors, s)
// 	return &s, nil
// }

// func (sh *SensorHub) refreshSensors() {
// 	for _, s := range sh.sensors {
// 		if s.Expired() {
// 			logger.Debug(fmt.Sprintf("refreshing %v", s))
// 			s.Refresh()
// 		}
// 	}
// }

// func (sh *SensorHub) Start() {
// 	sh.jobTicker = time.NewTicker(time.Duration(sh.config.SensorHubJobPeriod) * time.Second)
// 	go func() {
// 		for range sh.jobTicker.C {
// 			sh.refreshSensors()
// 		}
// 	}()
// }

// func (sh *SensorHub) Stop() {
// 	sh.jobTicker.Stop()
// }

// func (sh SensorHub) GetMetrics() []Metric {
// 	metrics := []Metric{}
// 	for _, s := range sh.sensors {
// 		metrics = append(metrics, s.GetMetrics()...)
// 	}
// 	return metrics
// }
