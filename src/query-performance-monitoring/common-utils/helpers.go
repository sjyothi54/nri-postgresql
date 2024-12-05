package common_utils

import (
	"fmt"
	"github.com/newrelic/infra-integrations-sdk/v3/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/v3/data/metric"
	"github.com/newrelic/infra-integrations-sdk/v3/integration"
	"github.com/newrelic/nri-postgresql/src/args"
	"reflect"
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
		var numericValue float64
		switch v := value.(type) {
		case int:
			numericValue = float64(v)
		case int64:
			numericValue = float64(v)
		case float64:
			numericValue = v
		default:
			fmt.Println("Error: gauge metric requires a numeric value")
			return
		}
		err := metricSet.SetMetric(name, numericValue, metric.GAUGE)
		if err != nil {
			fmt.Println("Error in setting metric1", err)
			return
		}
	case `attribute`:
		err := metricSet.SetMetric(name, fmt.Sprintf("%v", value), metric.ATTRIBUTE)
		if err != nil {
			fmt.Println("Error in setting metric", err)
			return
		}
	default:
		err := metricSet.SetMetric(name, value, metric.ATTRIBUTE)
		if err != nil {
			fmt.Println("Error in setting metric", err)
			return
		}
	}
}

func SetMetricsParser(instanceEntity *integration.Entity, eventName string, args args.ArgumentList, model interface{}) {
	metricSetIngestion := CreateMetricSet(instanceEntity, eventName, args)
	modelValue := reflect.ValueOf(model)
	modelType := reflect.TypeOf(model)
	for i := 0; i < modelValue.NumField(); i++ {
		field := modelValue.Field(i)
		fmt.Print("fieldooooo", field)
		fieldType := modelType.Field(i)
		metricName := fieldType.Tag.Get("metric_name")
		sourceType := fieldType.Tag.Get("source_type")

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			fmt.Println("heyyy")
			SetMetric(metricSetIngestion, metricName, field.Elem().Interface(), sourceType)
		} else if field.Kind() != reflect.Ptr {
			fmt.Println("Byeee")
			SetMetric(metricSetIngestion, metricName, field.Interface(), sourceType)
		}
	}
}
