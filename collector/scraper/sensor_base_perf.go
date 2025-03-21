package scraper

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/vmware/govmomi/vim25/types"
)

type BasePerfSensor struct {
	perfMetrics map[types.ManagedObjectReference]*MetricQueue
	scraper     *VCenterScraper
	metrics     struct {
		QueryTime      *SensorMetricDuration
		ClientWaitTime *SensorMetricDuration
		Status         *SensorMetricStatus
	}
	perfOptions []PerfOption
}

func (s *BasePerfSensor) Clean(maxAge time.Duration, logger *slog.Logger) {
	now := time.Now()
	total := 0
	for _, pm := range s.perfMetrics {
		ExpiredMetrics := pm.PopOlderOrEqualThan(now.Add(-maxAge))
		total = total + len(ExpiredMetrics)
	}
	if total > 0 {
		logger.Warn(fmt.Sprintf("Removed %d host-metrics which were not yet pulled", total))
	}
}

func (s *BasePerfSensor) PopAll(ref types.ManagedObjectReference) []*Metric {
	if perfMetrics, ok := s.perfMetrics[ref]; ok {
		return perfMetrics.PopAll()
	}
	return nil
}

func (s *BasePerfSensor) GetAllJsons() (map[string][]byte, error) {
	result := map[string][]byte{}

	for ref, queue := range s.perfMetrics {
		name := ref.Value
		jsonBytes, err := json.MarshalIndent(queue, "", "  ")
		if err != nil {
			return nil, err
		}
		result[name] = jsonBytes
	}
	return result, nil
}
