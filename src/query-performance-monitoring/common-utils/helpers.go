package common_utils

import (
	"errors"
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"math"
	"reflect"
	"strconv"
)

func CreateMetricSet(e *integration.Entity, sampleName string, args args.ArgumentList) *metric.Set {
	return metricSet(
		e,
		sampleName,
		args.Hostname,
		args.Port,
	)
}

func metricSet(e *integration.Entity, eventType, hostname string, port string) *metric.Set {
	return e.NewMetricSet(
		eventType,
		attribute.Attr("hostname", hostname),
		attribute.Attr("port", port),
	)
}

func SetMetric(metricSet *metric.Set, name string, value interface{}, sourceType string) {
	switch sourceType {
	case `gauge`:
		//var numericValue float64
		//switch v := value.(type) {
		//case int:
		//	numericValue, _ = castToFloat(v)
		//case int64:
		//	numericValue, _ = castToFloat(v)
		//case float64:
		//	numericValue = v
		//default:
		//	fmt.Println("Error: gauge metric requires a numeric value")
		//	return
		//}
		//fmt.Println("Numeric value: ", numericValue)
		err := metricSet.SetMetric(name, value, metric.GAUGE)
		if err != nil {
			fmt.Println("Error in setting metric1", err)
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			fmt.Println("Error in setting metric", err)
			return
		}
	default:
		fmt.Println("Error: metric type not supported")
		return

	}
}

func SetMetricsParser(instanceEntity *integration.Entity, eventName string, args args.ArgumentList, model interface{}) {
	metricSetIngestion := CreateMetricSet(instanceEntity, eventName, args)
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)
	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			SetMetric(metricSetIngestion, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			SetMetric(metricSetIngestion, metricName, field.Interface(), sourceType)
		}
	}
}

var ErrNonNumeric = errors.New("non-numeric value")

func castToFloat(value interface{}) (float64, error) {
	if b, ok := value.(bool); ok {
		if b {
			return 1, nil
		}
		return 0, nil
	}

	parsedValue, err := strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	if err != nil {
		return 0, err
	}

	if isNaNOrInf(parsedValue) {
		return 0, ErrNonNumeric
	}

	return parsedValue, nil
}

// isNaNOrInf checks if a float64 value is NaN or Infinity.
func isNaNOrInf(f float64) bool {
	return math.IsNaN(f) || math.IsInf(f, 0)
}
