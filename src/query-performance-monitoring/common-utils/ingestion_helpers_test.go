package commonutils

import (
	"errors"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"github.com/stretchr/testify/assert"
)

type mockMetricSet struct {
	metric.Set
	err error
}

func (m *mockMetricSet) SetMetric(name string, value interface{}, sourceType metric.SourceType) error {
	return m.err
}

func Test_SetMetric(t *testing.T) {
	tests := []struct {
		name       string
		sourceType string
		err        error
	}{
		{"GaugeError", "gauge", errors.New("gauge error")},
		{"AttributeError", "attribute", errors.New("attribute error")},
		{"DefaultError", "unknown", errors.New("default error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSet := &mockMetricSet{
				Set: metric.Set{
					Metrics: make(map[string]interface{}),
				},
				err: tt.err,
			}
			SetMetric(&mockSet.Set, "testMetric", 123, tt.sourceType)
			assert.Equal(t, tt.err, mockSet.err)
		})
	}
}

func Test_IngestMetric(t *testing.T) {
	pgIntegration, err := integration.New("test", "1.0.0")
	assert.NoError(t, err)

	metricList := []interface{}{
		struct {
			One int `metric_name:"one" source_type:"gauge" ingest_data:"true"`
		}{One: 1},
	}

	args := args.ArgumentList{
		Hostname: "localhost",
		Port:     "5432",
	}

	IngestMetric(metricList, "TestEvent", pgIntegration, args)
}
