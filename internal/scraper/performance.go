package scraper

import (
	"time"

	"github.com/sanderdescamps/govc_exporter/internal/timequeue"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/constraints"
)

type Metric struct {
	timequeue.Event
	Ref       types.ManagedObjectReference `json:"ref"`
	Name      string                       `json:"name"`
	Unit      string                       `json:"unit"`
	Instance  string                       `json:"instance"`
	Value     float64                      `json:"value"`
	TimeStamp time.Time                    `json:"time_stamp"`
}

func (m Metric) GetTimestamp() time.Time {
	return m.TimeStamp
}

func EntityMetricToMetric(entiry performance.EntityMetric) []Metric {
	result := []Metric{}
	for id, sample := range entiry.SampleInfo {
		for _, entityValue := range entiry.Value {
			result = append(result,
				Metric{
					Ref:      entiry.Entity,
					Name:     entityValue.Name,
					Unit:     entityValue.Unit,
					Instance: entityValue.Instance,
					Value: func() float64 {
						if entityValue.Value != nil {
							if len(entiry.SampleInfo) == len(entityValue.Value) {
								return float64(entityValue.Value[id])
							} else {
								return Avg(entityValue.Value)
							}
						}
						return 0
					}(),
					TimeStamp: sample.Timestamp,
				},
			)
		}

	}
	return result
}

type perfQuery struct {
	metrics        []string
	instance       string
	sampleCount    int32
	sampleInterval int32
	window         time.Duration
}

func (pq perfQuery) ToSpec() types.PerfQuerySpec {
	endtime := time.Now()
	starttime := endtime.Add(-pq.window)
	return types.PerfQuerySpec{
		MaxSample: pq.sampleCount,
		MetricId:  []types.PerfMetricId{{Instance: pq.instance}},
		// IntervalId: pq.sampleInterval,
		StartTime: &starttime,
		EndTime:   &endtime,
	}
}

type PerfOption func(*perfQuery)

func NewPerfQuery(options ...PerfOption) *perfQuery {
	result := &perfQuery{
		metrics:     []string{},
		instance:    "*",
		sampleCount: 1,
		// sampleInterval: 20,
	}
	for _, o := range options {
		o(result)
	}
	return result
}

func SetSamples[T constraints.Integer](num T) PerfOption {
	return func(pq *perfQuery) {
		pq.sampleCount = int32(num)
	}
}

func SetInterval(d time.Duration) PerfOption {
	return func(pq *perfQuery) {
		pq.sampleInterval = int32(d.Seconds())
	}
}

func SetWindow(d time.Duration) PerfOption {
	return func(pq *perfQuery) {
		pq.window = d
	}
}

func AddMetrics(m ...string) PerfOption {
	return func(pq *perfQuery) {
		pq.metrics = append(pq.metrics, m...)
	}
}

func SetMetrics(m ...string) PerfOption {
	return func(pq *perfQuery) {
		pq.metrics = m
	}
}

func AllHostMetrics() PerfOption {
	return SetMetrics([]string{
		"cpu.capacity.contention.average", "cpu.capacity.demand.average",
		"cpu.capacity.provisioned.average", "cpu.capacity.usage.average",
		"cpu.corecount.contention.average", "cpu.corecount.provisioned.average",
		"cpu.corecount.usage.average", "cpu.costop.summation", "cpu.demand.average",
		"cpu.entitlement.latest", "cpu.latency.average", "cpu.maxlimited.summation",
		"cpu.readiness.average", "cpu.ready.summation", "cpu.reservedCapacity.average",
		"cpu.usage.average", "cpu.usage.maximum", "cpu.usage.minimum",
		"cpu.usagemhz.average", "cpu.usagemhz.maximum", "cpu.usagemhz.minimum",
		"datastore.numberReadAveraged.average", "datastore.numberWriteAveraged.average",
		"datastore.read.average", "datastore.totalReadLatency.average",
		"datastore.totalWriteLatency.average", "datastore.write.average",
		"disk.throughput.contention.average", "disk.throughput.usage.average",
		"mem.active.average", "mem.active.maximum", "mem.active.minimum",
		"mem.capacity.contention.average", "mem.capacity.entitlement.average",
		"mem.capacity.provisioned.average", "mem.capacity.usable.average",
		"mem.capacity.usage.average", "mem.compressed.average", "mem.compressionRate.average",
		"mem.consumed.average", "mem.consumed.maximum", "mem.consumed.minimum",
		"mem.decompressionRate.average", "mem.entitlement.average", "mem.granted.average",
		"mem.granted.maximum", "mem.granted.minimum", "mem.overhead.average",
		"mem.overhead.maximum", "mem.overhead.minimum", "mem.reservedCapacity.average",
		"mem.shared.average", "mem.shared.maximum", "mem.shared.minimum",
		"mem.swapped.average", "mem.swapused.average", "mem.swapused.maximum",
		"mem.swapused.minimum", "mem.usage.average", "mem.usage.maximum",
		"mem.usage.minimum", "mem.vmmemctl.average", "mem.vmmemctl.maximum",
		"mem.vmmemctl.minimum", "mem.zero.average", "mem.zero.maximum",
		"mem.zero.minimum", "net.bytesRx.average", "net.bytesTx.average",
		"net.droppedRx.summation", "net.droppedTx.summation", "net.errorsRx.summation",
		"net.errorsTx.summation", "net.throughput.contention.summation",
		"net.throughput.provisioned.average", "net.throughput.usable.average",
		"net.throughput.usage.average", "power.energy.summation", "power.power.average",
		"power.powerCap.average", "sys.uptime.latest", "vmop.numChangeDS.latest",
		"vmop.numChangeHost.latest", "vmop.numChangeHostDS.latest", "vmop.numClone.latest",
		"vmop.numCreate.latest", "vmop.numDeploy.latest", "vmop.numDestroy.latest",
		"vmop.numPoweroff.latest", "vmop.numPoweron.latest", "vmop.numRebootGuest.latest",
		"vmop.numReconfigure.latest", "vmop.numRegister.latest", "vmop.numReset.latest",
		"vmop.numSVMotion.latest", "vmop.numShutdownGuest.latest",
		"vmop.numStandbyGuest.latest", "vmop.numSuspend.latest",
		"vmop.numUnregister.latest", "vmop.numVMotion.latest", "vmop.numXVMotion.latest",
	}...)
}

func SetInstance(i string) PerfOption {
	return func(pq *perfQuery) {
		pq.instance = i
	}
}
